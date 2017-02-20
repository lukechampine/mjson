// Package mjson modifies JSON super fast. Modifications consist of a path and
// a replacement value. A "path" in this context means a set of accessors
// (object keys or array indices) that reference an object or array element.
// For example, given this object:
//
//    {
//        "foo": {
//            "bars": [
//                {"baz":3}
//            ]
//        }
//    }
//
// The following path accesses the value "3":
//
//    foo.bars.0.baz
//
// The Set functions do nothing if the supplied path is malformed. A path is
// considered malformed if its path references an element that does not exist,
// including out-of-bound indices and object keys that are not valid JSON
// strings.
//
// As a special case, the length of the array (at application time) is a valid
// array index. When this index is the last accessor in the path, the value
// will be appended to the end of the array. If this special index is not the
// last accessor, the path is considered malformed (and thus is ignored).
package mjson

import (
	"bytes"
	gojson "encoding/json"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Set replaces the value at path in json with obj. If path is malformed, the
// original json is returned. If obj cannot be marshaled, Set panics.
func Set(json []byte, path string, obj interface{}) []byte {
	return rewritePath(json, path, marshal(obj), false)
}

// SetInPlace replaces the value at path in json with obj. If the length of
// obj is less than the existing value at that path, json will be modified in
// place. The result may contain extra whitespace. If path is malformed, the
// original json is returned. If obj cannot be marshaled, SetInPlace panics.
func SetInPlace(json []byte, path string, obj interface{}) []byte {
	return rewritePath(json, path, marshal(obj), true)
}

// SetRawInPlace replaces the value at path in json with val. If the length of
// val is less than the existing value at that path, json will be modified in
// place. The result may contain extra whitespace. If path is malformed, the
// original json is returned. If val cannot be marshaled, SetRawInPlace
// panics.
func SetRawInPlace(json []byte, path string, val []byte) []byte {
	return rewritePath(json, path, val, true)
}

// rewritePath replaces the value at path in json with val. If inPlace is
// true, the returned slice may share underlying memory with json. If path is
// malformed, the original json is returned.
func rewritePath(json []byte, path string, val []byte, inPlace bool) []byte {
	if path == "" {
		if inPlace {
			return append(json[:0], val...)
		}
		return append([]byte(nil), val...)
	}

	var lastAcc string
	var i int
	for j := 0; lastAcc == ""; j++ {
		// determine next accessor by seeking to .
		dotIndex := strings.IndexByte(path[j:], '.')
		if dotIndex == -1 {
			// not found; this is the last accessor
			dotIndex = len(path[j:])
			lastAcc = path[j:]
		}
		acc := path[j : j+dotIndex]
		j += dotIndex

		// seek to accessor
		accIndex := locateAccessor(json[i:], acc)
		if accIndex == -1 {
			// not found; return unmodified
			return json
		} else if (json[accIndex] == ']' || json[accIndex] == '}' || json[accIndex] == 'l') && lastAcc == "" {
			// only the last accessor may append
			return json
		}
		i += accIndex
	}
	// hack for appending to null
	appendNull := false
	if json[i] == 'l' && lastAcc == "0" {
		i -= 3
		appendNull = true
	}

	rest := consumeValue(json[i:])
	if inPlace {
		// can we replace without allocating?
		oldLen, newLen := 0, len(val)
		if json[i] != '}' && json[i] != ']' {
			oldLen = len(json[i:]) - len(rest)
		}
		if appendNull {
			newLen += 2 // account for []
		}
		if newLen <= oldLen {
			// new val is smaller; rewrite in-place
			if appendNull {
				json[i] = '['
				copy(json[i+1:], val)
				json[i+newLen-1] = ']'
			} else {
				copy(json[i:], val)
			}
			i += newLen
			for j := 0; j < oldLen-newLen; j++ {
				json[i+j] = ' ' // pad with whitespace
			}
			return json
		}
	}

	// replace old value
	newJSON := make([]byte, 0, len(json)+len(val)+len(lastAcc)) // reasonable guess
	newJSON = append(newJSON, json[:i]...)
	switch json[i] {
	default:
		newJSON = append(newJSON, val...)

	case '}': // insert a new key
		if prevChar(json, i) != '{' {
			// if the object is not empty, insert an extra ,
			newJSON = append(newJSON, ',')
		}
		// insert key
		newJSON = append(newJSON, '"')
		newJSON = append(newJSON, lastAcc...)
		newJSON = append(newJSON, '"', ':')
		newJSON = append(newJSON, val...)

	case ']': // append to an array
		if prevChar(json, i) != '[' {
			// if the array is not empty, insert an extra ,
			newJSON = append(newJSON, ',')
		}
		newJSON = append(newJSON, val...)

	case 'n': // replace null with a single-element array
		newJSON = append(newJSON, '[')
		newJSON = append(newJSON, val...)
		newJSON = append(newJSON, ']')
	}
	newJSON = append(newJSON, rest...)
	return newJSON
}

// locateAccessor returns the offset of acc in json.
func locateAccessor(json []byte, acc string) int {
	origLen := len(json)
	json = consumeWhitespace(json)
	if len(json) == 0 || len(json) < len(acc) {
		return -1
	}

	// acc must refer to either an object key or an array index. So if we
	// don't see a { or [, the path is invalid. We also allow a special case
	// for null -- it is treated as the empty array [].
	switch json[0] {
	default:
		return -1

	case '{': // object
		bacc := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: (*reflect.StringHeader)(unsafe.Pointer(&acc)).Data,
			Len:  len(acc),
			Cap:  len(acc),
		}))
		json = consumeSeparator(json) // consume {
		// iterate through keys, searching for acc
		for json[0] != '}' {
			var key []byte
			key, json = parseString(json)
			json = consumeWhitespace(json)
			json = consumeSeparator(json) // consume :
			if bytes.Equal(key, bacc) {
				// acc found
				return origLen - len(json)
			}
			json = consumeValue(json)
			json = consumeWhitespace(json)
			if json[0] == ',' {
				json = consumeSeparator(json) // consume ,
			}
		}
		// acc not found; return the offset of the closing }
		return origLen - len(json)

	case '[': // array
		// is accessor possibly an array index?
		n, err := strconv.Atoi(acc)
		if err != nil || n < 0 {
			// invalid index
			return -1
		}
		json = consumeSeparator(json) // consume [
		// consume n keys, stopping early if we hit the end of the array
		var arrayLen int
		for n > arrayLen && json[0] != ']' {
			json = consumeValue(json)
			arrayLen++
			json = consumeWhitespace(json)
			if json[0] == ',' {
				json = consumeSeparator(json) // consume ,
			}
		}
		if n > arrayLen {
			// Note that n == arrayLen is allowed. In this case, an append
			// operation is desired; we return the offset of the closing ].
			return -1
		}
		return origLen - len(json)

	case 'n': // null -- interpreted as []
		// acc must be 0 to append to null
		if n, err := strconv.Atoi(acc); err != nil || n != 0 {
			return -1
		}
		// return the offset of l
		json = json[3:]
		return origLen - len(json)
	}
}

func parseString(json []byte) ([]byte, []byte) {
	after := consumeString(json)
	strLen := len(json) - len(after) - 2
	return json[1 : 1+strLen], after
}

func consumeWhitespace(json []byte) []byte {
	for i := range json {
		if c := json[i]; c > ' ' || (c != ' ' && c != '\t' && c != '\n' && c != '\r') {
			return json[i:]
		}
	}
	return json[len(json):]
}

func prevChar(json []byte, i int) byte {
	// seek backwards from i until we hit a non-whitespace char
	for j := i - 1; j >= 0; j-- {
		if c := json[j]; c > ' ' || (c != ' ' && c != '\t' && c != '\n' && c != '\r') {
			return json[j]
		}
	}
	return json[i]
}

func consumeSeparator(json []byte) []byte {
	json = json[1:] // consume one of [ { } ] : ,
	return consumeWhitespace(json)
}

func consumeValue(json []byte) []byte {
	// determine value type
	switch json[0] {
	case '{': // object
		return consumeObject(json)
	case '[': // array
		return consumeArray(json)
	case '"': // string
		return consumeString(json)
	case 't', 'n': // true or null
		return json[4:]
	case 'f': // false
		return json[5:]
	default: // number
		return consumeNumber(json)
	}
}

func consumeObject(json []byte) []byte {
	json = json[1:] // consume {
	// seek to next {, }, or ". Each time we encounter a {, increment n. Each
	// time encounter a }, decrement n. Exit when n == 0. If we encounter ",
	// consume the string.
	n := 1
	for n > 0 {
		json = json[indexObject(json):]
		switch json[0] {
		case '{':
			n++
			json = json[1:] // consume {
		case '}':
			n--
			json = json[1:] // consume }
		case '"':
			json = consumeString(json)
		}
	}
	return json
}

func indexObject(json []byte) int {
	for i, c := range json {
		if c == '{' || c == '}' || c == '"' {
			return i
		}
	}
	return -1
}

func indexArray(json []byte) int {
	for i, c := range json {
		if c == '[' || c == ']' || c == '"' {
			return i
		}
	}
	return -1
}

func consumeArray(json []byte) []byte {
	json = json[1:] // consume [
	// seek to next [, ], or ". Each time we encounter a [, increment n. Each
	// time encounter a ], decrement n. Exit when n == 0. If we encounter ",
	// consume the string.
	n := 1
	for n > 0 {
		json = json[indexArray(json):]
		switch json[0] {
		case '[':
			n++
			json = json[1:] // consume [
		case ']':
			n--
			json = json[1:] // consume ]
		case '"':
			json = consumeString(json)
		}
	}
	return json
}

func consumeString(json []byte) []byte {
	for i := 1; i < len(json); i++ {
		if json[i] == '"' {
			return json[i+1:]
		}
		if json[i] == '\\' {
			i++
		}
	}
	return json
}

func consumeNumber(json []byte) []byte {
	for i, c := range json {
		if !('0' <= c && c <= '9') {
			switch c {
			case '+', '-', '.', 'e', 'E':
			default:
				return json[i:]
			}
		}
	}
	return json[len(json):]
}

// marshal marshals obj as JSON. If obj has a MarshalJSON method, it is called
// directly. Note that this may produce invalid JSON. If obj cannot be
// marshaled, marshal panics.
func marshal(obj interface{}) []byte {
	if m, ok := obj.(gojson.Marshaler); ok {
		b, err := m.MarshalJSON()
		if err != nil {
			panic(err)
		}
		return b
	}

	switch v := obj.(type) {
	default:
		b, err := gojson.Marshal(obj)
		if err != nil {
			panic(err)
		}
		return b

	case int:
		return strconv.AppendInt(nil, int64(v), 10)
	case int8:
		return strconv.AppendInt(nil, int64(v), 10)
	case int16:
		return strconv.AppendInt(nil, int64(v), 10)
	case int32:
		return strconv.AppendInt(nil, int64(v), 10)
	case int64:
		return strconv.AppendInt(nil, int64(v), 10)
	case uint:
		return strconv.AppendUint(nil, uint64(v), 10)
	case uint8:
		return strconv.AppendUint(nil, uint64(v), 10)
	case uint16:
		return strconv.AppendUint(nil, uint64(v), 10)
	case uint32:
		return strconv.AppendUint(nil, uint64(v), 10)
	case uint64:
		return strconv.AppendUint(nil, uint64(v), 10)
	case float32:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64)
	case float64:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64)
	case string:
		return strconv.AppendQuote(nil, v)
	case bool:
		if v {
			return []byte("true")
		} else {
			return []byte("false")
		}
	}
}
