package mjson

import (
	"bytes"
	"testing"

	"github.com/tidwall/sjson"
)

func TestSet(t *testing.T) {
	tests := []struct {
		json string
		path string
		val  interface{}
		exp  string
	}{
		{``, ``, "foo", `"foo"`},
		{`"foo"`, ``, "bar", `"bar"`},
		// object
		{`{"foo":"bar"}`, `foo`, "baz", `{"foo":"baz"}`},
		{`{"foo":"bar", "bar":"baz"}`, `bar`, "quux", `{"foo":"bar", "bar":"quux"}`},
		{`{"foo": {"bar": "baz"}}`, `foo.bar`, false, `{"foo": {"bar": false}}`},
		{`{"foo":"bar"}`, `bar`, "baz", `{"foo":"bar","bar":"baz"}`},
		{`{"foo": {}}`, `foo.bar`, 3, `{"foo": {"bar":3}}`},
		// array
		{`[]`, `foo`, "bar", `[]`},
		{`[1]`, `0`, "bar", `["bar"]`},
		{`[1, 2]`, `0`, "bar", `["bar", 2]`},
		{`[1, 2]`, `1`, nil, `[1, null]`},
		{`[]`, `0`, "bar", `["bar"]`},
		{`[1]`, `1`, 3.1, `[1,3.1]`},
		{`["foo", "bar"]`, `2`, "baz", `["foo", "bar","baz"]`},
		{`[[1,2], [3,4]]`, `1.0`, "baz", `[[1,2], ["baz",4]]`},
		{`[[1,2], [3,4]]`, `1.2`, "baz", `[[1,2], [3,4,"baz"]]`},
		{`[[1,2], [3,4]]`, `2.0`, "baz", `[[1,2], [3,4]]`},
		{`[[1,2], [3,4]]`, `2`, "baz", `[[1,2], [3,4],"baz"]`},
		// array in object
		{`{"foo": [1,2]}`, `foo.0`, "bar", `{"foo": ["bar",2]}`},
		{`{"foo": [1,2]}`, `foo.2`, "bar", `{"foo": [1,2,"bar"]}`},
		{`{"foo": [1,2]}`, `bar.2`, "bar", `{"foo": [1,2]}`},
		// object in array
		{`[{"foo": "bar"}]`, `0.foo`, "baz", `[{"foo": "baz"}]`},
		{`[{}, {"foo": "bar"}]`, `1.foo`, "baz", `[{}, {"foo": "baz"}]`},
		{`[{"foo": "bar"}]`, `1.foo`, `"baz"`, `[{"foo": "bar"}]`},
		// null
		{`null`, `foo`, "bar", `null`},
		{`null`, `0`, "bar", `["bar"]`},
		{`{"foo": null}`, `foo.0`, "bar", `{"foo": ["bar"]}`},
		// monster
		{`{"foo": [{}, {"bar": [{"baz":""}]}}]`, `foo.1.bar.0.baz`, "quux", `{"foo": [{}, {"bar": [{"baz":"quux"}]}}]`},
	}
	for _, test := range tests {
		if res := Set([]byte(test.json), test.path, test.val); string(res) != test.exp {
			t.Errorf("Set('%s', %q, '%s'): expected '%s', got '%s'", test.json, test.path, test.val, test.exp, res)
		}
	}
}

func TestSetInPlace(t *testing.T) {
	tests := []struct {
		json string
		path string
		val  interface{}
		exp  string
	}{
		{`"foo"`, ``, "", `""`},
		{`"foo"`, ``, "bar", `"bar"`},
		// object
		{`{"foo":"bar"}`, `foo`, "baz", `{"foo":"baz"}`},
		{`{"foo":"bar", "bar":"quux"}`, `bar`, "baz", `{"foo":"bar", "bar":"baz" }`},
		{`{"foo": {"bar": "baz"}}`, `foo.bar`, 1, `{"foo": {"bar": 1    }}`},
		{`{"foo":"bar"}`, `bar`, "baz", `{"foo":"bar","bar":"baz"}`},
		{`{"foo": {}}`, `foo.bar`, "baz", `{"foo": {"bar":"baz"}}`},
		// array
		{`[]`, `foo`, "bar", `[]`},
		{`[10]`, `0`, 1, `[1 ]`},
		{`[1, 2]`, `0`, "bar", `["bar", 2]`},
		{`[[1,2], [3,4]]`, `1`, "baz", `[[1,2], "baz"]`},
		// null
		{`null`, `foo`, "bar", `null`},
		{`null`, `0`, 1, `[1] `},
		{`{"foo": null}`, `foo.0`, 1, `{"foo": [1] }`},
	}
	for _, test := range tests {
		if res := SetInPlace([]byte(test.json), test.path, test.val); string(res) != test.exp {
			t.Errorf("SetInPlace('%s', %q, '%s'): expected '%s', got '%s'", test.json, test.path, test.val, test.exp, res)
		}
	}
}

func TestRewritePath(t *testing.T) {
	tests := []struct {
		json string
		path string
		val  string
		exp  string
	}{
		{``, ``, ``, ``},
		{`"foo"`, ``, ``, ``},
		{``, `foo`, ``, ``},
		{``, ``, `"foo"`, `"foo"`},
		{`"foo"`, ``, `"bar"`, `"bar"`},
		// object
		{`{"foo":"bar"}`, `foo`, `"baz"`, `{"foo":"baz"}`},
		{`{"foo":"bar", "bar":"baz"}`, `bar`, `"quux"`, `{"foo":"bar", "bar":"quux"}`},
		{`{"foo": {"bar": "baz"}}`, `foo.bar`, `"quux"`, `{"foo": {"bar": "quux"}}`},
		{`{"foo":"bar"}`, `bar`, `"baz"`, `{"foo":"bar","bar":"baz"}`},
		{`{"foo": {}}`, `foo.bar`, `"baz"`, `{"foo": {"bar":"baz"}}`},
		// array
		{`[]`, `foo`, `"bar"`, `[]`},
		{`[1]`, `0`, `"bar"`, `["bar"]`},
		{`[1, 2]`, `0`, `"bar"`, `["bar", 2]`},
		{`[1, 2]`, `1`, `"bar"`, `[1, "bar"]`},
		{`[]`, `0`, `"bar"`, `["bar"]`},
		{`[1]`, `1`, `"bar"`, `[1,"bar"]`},
		{`["foo", "bar"]`, `2`, `"baz"`, `["foo", "bar","baz"]`},
		{`[[1,2], [3,4]]`, `1.0`, `"baz"`, `[[1,2], ["baz",4]]`},
		{`[[1,2], [3,4]]`, `1.2`, `"baz"`, `[[1,2], [3,4,"baz"]]`},
		{`[[1,2], [3,4]]`, `2.0`, `"baz"`, `[[1,2], [3,4]]`},
		{`[[1,2], [3,4]]`, `2`, `"baz"`, `[[1,2], [3,4],"baz"]`},
		// array in object
		{`{"foo": [1,2]}`, `foo.0`, `"bar"`, `{"foo": ["bar",2]}`},
		{`{"foo": [1,2]}`, `foo.2`, `"bar"`, `{"foo": [1,2,"bar"]}`},
		{`{"foo": [1,2]}`, `bar.2`, `"bar"`, `{"foo": [1,2]}`},
		// object in array
		{`[{"foo": "bar"}]`, `0.foo`, `"baz"`, `[{"foo": "baz"}]`},
		{`[{}, {"foo": "bar"}]`, `1.foo`, `"baz"`, `[{}, {"foo": "baz"}]`},
		{`[{"foo": "bar"}]`, `1.foo`, `"baz"`, `[{"foo": "bar"}]`},
		// null
		{`null`, `foo`, `"bar"`, `null`},
		{`null`, `0`, `"bar"`, `["bar"]`},
		{`{"foo": null}`, `foo.0`, `"bar"`, `{"foo": ["bar"]}`},
		// monster
		{`{"foo": [{}, {"bar": [{"baz":""}]}}]`, `foo.1.bar.0.baz`, `"quux"`, `{"foo": [{}, {"bar": [{"baz":"quux"}]}}]`},
	}
	for _, test := range tests {
		if res := rewritePath([]byte(test.json), test.path, []byte(test.val), false); string(res) != test.exp {
			t.Errorf("rewritePath('%s', %q, '%s'): expected '%s', got '%s'", test.json, test.path, test.val, test.exp, res)
		}
	}
}

func TestRewritePathInPlace(t *testing.T) {
	tests := []struct {
		json string
		path string
		val  string
		exp  string
	}{
		{`"foo"`, ``, ``, ``},
		{`"foo"`, ``, `"bar"`, `"bar"`},
		// object
		{`{"foo":"bar"}`, `foo`, `"baz"`, `{"foo":"baz"}`},
		{`{"foo":"bar", "bar":"quux"}`, `bar`, `"baz"`, `{"foo":"bar", "bar":"baz" }`},
		{`{"foo": {"bar": "baz"}}`, `foo.bar`, `"quux"`, `{"foo": {"bar": "quux"}}`},
		{`{"foo":"bar"}`, `bar`, `"baz"`, `{"foo":"bar","bar":"baz"}`},
		{`{"foo": {}}`, `foo.bar`, `"baz"`, `{"foo": {"bar":"baz"}}`},
		// array
		{`[]`, `foo`, `"bar"`, `[]`},
		{`[10]`, `0`, `1`, `[1 ]`},
		{`[1, 2]`, `0`, `"bar"`, `["bar", 2]`},
		{`[[1,2], [3,4]]`, `1`, `"baz"`, `[[1,2], "baz"]`},
		// null
		{`null`, `foo`, `"bar"`, `null`},
		{`null`, `0`, `1`, `[1] `},
		{`{"foo": null}`, `foo.0`, `1`, `{"foo": [1] }`},
	}
	for _, test := range tests {
		if res := rewritePath([]byte(test.json), test.path, []byte(test.val), true); string(res) != test.exp {
			t.Errorf("rewritePath('%s', %q, '%s'): expected '%s', got '%s'", test.json, test.path, test.val, test.exp, res)
		}
	}
}

func TestLocateAccessor(t *testing.T) {
	tests := []struct {
		json string
		acc  string
		loc  int
	}{
		// object
		{`{}`, `foo`, -1},
		{`{"foo":0}`, `foo`, 7},
		{`{"foo":0}`, `bar`, 8}, // special case
		{`{"foo":0}3`, `foo`, 7},
		{`{"foo":0} 3`, `foo`, 7},
		{`{"foo":0,"bar":7}`, `bar`, len(`{"foo":0,"bar":`)},
		{`{"foo":0 , "bar":7}`, `bar`, len(`{"foo":0 , "bar":`)},
		{`{"foo":0,"bar":7}3`, `bar`, len(`{"foo":0,"bar":`)},
		{`{"foo":0,"bar":7} 3`, `bar`, len(`{"foo":0,"bar":`)},
		// array
		{`[1,2,3]`, `0`, 1},
		{`[1,2,3]`, `1`, 3},
		{`[1,2,3]`, `2`, 5},
		{`[1,2,3]`, `3`, 6}, // special case
		{`[1,2,3]`, `4`, -1},
		{`[1,2,3]`, `foo`, -1},
		{`[]`, `0`, 1},
		{`[]`, `1`, -1},
		// null
		{`null`, `0`, 3}, // special case
		{`null`, `1`, -1},
		// string
		{`"foo"`, `foo`, -1},
		{`"{\"foo\": 3}"`, `foo`, -1},
		// number
		{`3`, `foo`, -1},
		{`3`, `3`, -1},
	}
	for _, test := range tests {
		if loc := locateAccessor([]byte(test.json), test.acc); loc != test.loc {
			t.Errorf("locateAccessor('%s', %q): expected %v, got %v", test.json, test.acc, test.loc, loc)
		}
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		json string
		str  string
		rest string
	}{
		{`""`, ``, ``},
		{`"foo"`, `foo`, ``},
		{`"foo":"bar"`, `foo`, `:"bar"`},
		{`"foo" : "bar"`, `foo`, ` : "bar"`},
		{`"foo\"bar"`, `foo\"bar`, ``},
		{`"foo\"bar":"baz"`, `foo\"bar`, `:"baz"`},
		{`"foo\\\"bar":"baz"`, `foo\\\"bar`, `:"baz"`},
	}
	for _, test := range tests {
		if str, rest := parseString([]byte(test.json)); string(str) != test.str || string(rest) != test.rest {
			t.Errorf("parseString('%s'): expected (%q, '%s'), got (%q, '%s')", test.json, test.str, test.rest, str, rest)
		}
	}
}

func TestConsumeWhitespace(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		{" ", ``},
		{" 3", `3`},
		{"\t", ``},
		{"\t3", `3`},
		{"\n", ``},
		{"\n3", `3`},
		{"\r", ``},
		{"\r3", `3`},
		{" \t\n\r", ``},
		{" \t\n\r3", `3`},
	}
	for _, test := range tests {
		if rest := consumeWhitespace([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeWhitespace('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestPrevChar(t *testing.T) {
	tests := []struct {
		json  string
		index int
		char  byte
	}{
		{" ", 0, ' '},
		{" 3", 1, '3'},
		{"\t", 0, '\t'},
		{"\t3", 1, '3'},
		{"\n", 0, '\n'},
		{"\n3", 1, '3'},
		{"\r", 0, '\r'},
		{"\r3", 1, '3'},
		{" \t\n\r", 3, '\r'},
		{" \t\n\r3", 4, '3'},
		{"[]", 1, '['},
		{" ]", 1, ']'},
		{"[1]", 2, '1'},
		{"{}", 1, '{'},
		{" }", 1, '}'},
		{`{"foo"}`, 6, '"'},
	}
	for _, test := range tests {
		if char := prevChar([]byte(test.json), test.index); char != test.char {
			t.Errorf("prevChar(%q, %d): expected %q, got %q", test.json, test.index, test.char, char)
		}
	}
}

func TestConsumeSeparator(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		{"[", ``},
		{"[ \r\n\t", ``},
		{"[3", `3`},
		{"[ \r\n\t3", `3`},
		{"{", ``},
		{"{ \r\n\t", ``},
		{"{3", `3`},
		{"{ \r\n\t3", `3`},
		{"}", ``},
		{"} \r\n\t", ``},
		{"}3", `3`},
		{"} \r\n\t3", `3`},
		{"]", ``},
		{"] \r\n\t", ``},
		{"]3", `3`},
		{"] \r\n\t3", `3`},
		{":", ``},
		{": \r\n\t", ``},
		{":3", `3`},
		{": \r\n\t3", `3`},
		{",", ``},
		{", \r\n\t", ``},
		{",3", `3`},
		{", \r\n\t3", `3`},
	}
	for _, test := range tests {
		if rest := consumeSeparator([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeSeparator('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestConsumeValue(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		// object
		{`{}`, ``},
		{`{}3`, `3`},
		{`{} 3`, ` 3`},
		{`{"foo":0}`, ``},
		{`{"foo":0}3`, `3`},
		{`{"foo":0} 3`, ` 3`},
		{`{"foo":0,"bar":7}`, ``},
		{`{"foo":0,"bar":7}3`, `3`},
		{`{"foo":0,"bar":7} 3`, ` 3`},
		{`{"foo":0 , "bar":7}`, ``},
		{`{"foo":0 , "bar":7}3`, `3`},
		{`{"foo":0 , "bar":7} 3`, ` 3`},
		{`{"":""}`, ``},
		{`{"":""}3`, `3`},
		{`{"":""} 3`, ` 3`},
		{`{"}":"}"}`, ``},
		{`{"}":"}"}3`, `3`},
		{`{"}":"}"} 3`, ` 3`},
		{`{"":{}}`, ``},
		{`{"":{}}3`, `3`},
		{`{"":{}} 3`, ` 3`},
		{`{"}":["}{"]}`, ``},
		{`{"}":["}{"]}3`, `3`},
		{`{"}":["}{"]} 3`, ` 3`},
		{`{"":{"":{"":{}}}}`, ``},
		{`{"":{"":{"":{}}}}3`, `3`},
		{`{"":{"":{"":{}}}} 3`, ` 3`},
		// array
		{`[]`, ``},
		{`[]3`, `3`},
		{`[] 3`, ` 3`},
		{`[0]`, ``},
		{`[0]3`, `3`},
		{`[0] 3`, ` 3`},
		{`[0,1]`, ``},
		{`[0,1]3`, `3`},
		{`[0,1] 3`, ` 3`},
		{`[0 , 1]`, ``},
		{`[0 , 1]3`, `3`},
		{`[0 , 1] 3`, ` 3`},
		{`["", ""]`, ``},
		{`["", ""]3`, `3`},
		{`["", ""] 3`, ` 3`},
		{`[[], []]`, ``},
		{`[[], []]3`, `3`},
		{`[[], []] 3`, ` 3`},
		{`["[", "]"]`, ``},
		{`["[", "]"]3`, `3`},
		{`["[", "]"] 3`, ` 3`},
		{`[["]", "]"], [{"foo":[]}]]`, ``},
		{`[["]", "]"], [{"foo":[]}]]3`, `3`},
		{`[["]", "]"], [{"foo":[]}]] 3`, ` 3`},
		// string
		{`""`, ``},
		{`"foo"`, ``},
		{`"foo":"bar"`, `:"bar"`},
		{`"foo" : "bar"`, ` : "bar"`},
		{`"foo\"bar"`, ``},
		{`"foo\"bar":"baz"`, `:"baz"`},
		// true, false, null
		{`true`, ``},
		{`false`, ``},
		{`null`, ``},
		{`true3`, `3`},
		{`false3`, `3`},
		{`null3`, `3`},
		{`true 3`, ` 3`},
		{`false 3`, ` 3`},
		{`null 3`, ` 3`},
		// number
		{`-0`, ``},
		{`-0 true`, ` true`},
		{`0`, ``},
		{`0 true`, ` true`},
		{`0.0`, ``},
		{`0.0 true`, ` true`},
		{`1.0`, ``},
		{`1.0 true`, ` true`},
		{`10`, ``},
		{`10 true`, ` true`},
		{`10.1`, ``},
		{`10.1 true`, ` true`},
		{`1e7`, ``},
		{`1e7 true`, ` true`},
		{`1e+7`, ``},
		{`1e+7 true`, ` true`},
		{`1e-7`, ``},
		{`1e-7 true`, ` true`},
		{`1.0e7`, ``},
		{`1.0e7 true`, ` true`},
		{`1.0e+7`, ``},
		{`1.0e+7 true`, ` true`},
		{`1.0e-7`, ``},
		{`1.0e-7 true`, ` true`},
		{`10.1e7`, ``},
		{`10.1e7 true`, ` true`},
		{`10.1e+7`, ``},
		{`10.1e+7 true`, ` true`},
		{`10.1e-7`, ``},
		{`10.1e-7 true`, ` true`},
	}
	for _, test := range tests {
		if rest := consumeValue([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeValue('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestConsumeObject(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		{`{}`, ``},
		{`{}3`, `3`},
		{`{} 3`, ` 3`},
		{`{"foo":0}`, ``},
		{`{"foo":0}3`, `3`},
		{`{"foo":0} 3`, ` 3`},
		{`{"foo":0,"bar":7}`, ``},
		{`{"foo":0,"bar":7}3`, `3`},
		{`{"foo":0,"bar":7} 3`, ` 3`},
		{`{"foo":0 , "bar":7}`, ``},
		{`{"foo":0 , "bar":7}3`, `3`},
		{`{"foo":0 , "bar":7} 3`, ` 3`},
		{`{"":""}`, ``},
		{`{"":""}3`, `3`},
		{`{"":""} 3`, ` 3`},
		{`{"}":"}"}`, ``},
		{`{"}":"}"}3`, `3`},
		{`{"}":"}"} 3`, ` 3`},
		{`{"":{}}`, ``},
		{`{"":{}}3`, `3`},
		{`{"":{}} 3`, ` 3`},
		{`{"}":["}{"]}`, ``},
		{`{"}":["}{"]}3`, `3`},
		{`{"}":["}{"]} 3`, ` 3`},
		{`{"":{"":{"":{}}}}`, ``},
		{`{"":{"":{"":{}}}}3`, `3`},
		{`{"":{"":{"":{}}}} 3`, ` 3`},
	}
	for _, test := range tests {
		if rest := consumeObject([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeObject('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestConsumeArray(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		{`[]`, ``},
		{`[]3`, `3`},
		{`[] 3`, ` 3`},
		{`[0]`, ``},
		{`[0]3`, `3`},
		{`[0] 3`, ` 3`},
		{`[0,1]`, ``},
		{`[0,1]3`, `3`},
		{`[0,1] 3`, ` 3`},
		{`[0 , 1]`, ``},
		{`[0 , 1]3`, `3`},
		{`[0 , 1] 3`, ` 3`},
		{`["", ""]`, ``},
		{`["", ""]3`, `3`},
		{`["", ""] 3`, ` 3`},
		{`[[], []]`, ``},
		{`[[], []]3`, `3`},
		{`[[], []] 3`, ` 3`},
		{`["[", "]"]`, ``},
		{`["[", "]"]3`, `3`},
		{`["[", "]"] 3`, ` 3`},
		{`[["]", "]"], [{"foo":[]}]]`, ``},
		{`[["]", "]"], [{"foo":[]}]]3`, `3`},
		{`[["]", "]"], [{"foo":[]}]] 3`, ` 3`},
	}
	for _, test := range tests {
		if rest := consumeArray([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeArray('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestConsumeString(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		{`""`, ``},
		{`"foo"`, ``},
		{`"foo":"bar"`, `:"bar"`},
		{`"foo" : "bar"`, ` : "bar"`},
		{`"foo\"bar"`, ``},
		{`"foo\"bar":"baz"`, `:"baz"`},
		{`"foo\\\"bar":"baz"`, `:"baz"`},
	}
	for _, test := range tests {
		if rest := consumeString([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeString('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestConsumeNumber(t *testing.T) {
	tests := []struct {
		json string
		rest string
	}{
		{`-0`, ``},
		{`-0 true`, ` true`},
		{`0`, ``},
		{`0 true`, ` true`},
		{`0.0`, ``},
		{`0.0 true`, ` true`},
		{`1.0`, ``},
		{`1.0 true`, ` true`},
		{`10`, ``},
		{`10 true`, ` true`},
		{`10.1`, ``},
		{`10.1 true`, ` true`},
		{`1e7`, ``},
		{`1e7 true`, ` true`},
		{`1e+7`, ``},
		{`1e+7 true`, ` true`},
		{`1e-7`, ``},
		{`1e-7 true`, ` true`},
		{`1.0e7`, ``},
		{`1.0e7 true`, ` true`},
		{`1.0e+7`, ``},
		{`1.0e+7 true`, ` true`},
		{`1.0e-7`, ``},
		{`1.0e-7 true`, ` true`},
		{`10.1e7`, ``},
		{`10.1e7 true`, ` true`},
		{`10.1e+7`, ``},
		{`10.1e+7 true`, ` true`},
		{`10.1e-7`, ``},
		{`10.1e-7 true`, ` true`},
	}
	for _, test := range tests {
		if rest := consumeNumber([]byte(test.json)); string(rest) != test.rest {
			t.Errorf("consumeNumber('%s'): expected '%s', got '%s'", test.json, test.rest, rest)
		}
	}
}

func TestMarshal(t *testing.T) {
	b := marshal(3)
	if !bytes.Equal(b, []byte(`3`)) {
		t.Error("marshal mismatch:", string(b))
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	marshal(map[interface{}]interface{}{})
}

func BenchmarkConsumeWhitespace(b *testing.B) {
	json := []byte("  \n  \r \t  :")
	for i := 0; i < b.N; i++ {
		consumeWhitespace(json)
	}
}

func BenchmarkConsumeObject(b *testing.B) {
	json := []byte(`{"foo": {"bar": {"baz": "quux"}}}`)
	for i := 0; i < b.N; i++ {
		consumeObject(json)
	}
}

func BenchmarkConsumeString(b *testing.B) {
	json := []byte(`"{foo: {bar: {baz: quux}}}"`)
	for i := 0; i < b.N; i++ {
		consumeString(json)
	}
}

func BenchmarkConsumeNumber(b *testing.B) {
	json := []byte(`-314.1592653e-2`)
	for i := 0; i < b.N; i++ {
		consumeNumber(json)
	}
}

func BenchmarkParseString(b *testing.B) {
	json := []byte(`"{\"foo\": {\"bar\": {\"baz\": \"quux\"}}}"`)
	for i := 0; i < b.N; i++ {
		parseString(json)
	}
}

const benchJSON = `
{
  "widget": {
    "debug": "on",
    "window": {
      "title": "Sample Konfabulator Widget",
      "name": "main_window",
      "width": 500,
      "height": 500
    },
    "image": { 
      "src": "Images/Sun.png",
      "hOffset": 250,
      "vOffset": 250,
      "alignment": "center"
    },
    "text": {
      "data": "Click Here",
      "size": 36,
      "style": "bold",
      "vOffset": 100,
      "alignment": "center",
      "onMouseUp": "sun1.opacity = (sun1.opacity / 100) * 90;"
    }
  }
}    
`

var benchPaths = []string{
	"widget.window.name",
	"widget.image.hOffset",
	"widget.text.onMouseUp",
}

func Benchmark_Set(b *testing.B) {
	data := []byte(benchJSON)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				Set(data, path, "1")
			case "widget.image.hOffset":
				Set(data, path, 1)
			case "widget.text.onMouseUp":
				Set(data, path, "1")
			}
		}
	}
	b.N *= len(benchPaths)
}

func Benchmark_SetInPlace(b *testing.B) {
	data := []byte(benchJSON)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				SetInPlace(data, path, "1")
			case "widget.image.hOffset":
				SetInPlace(data, path, 1)
			case "widget.text.onMouseUp":
				SetInPlace(data, path, "1")
			}
		}
	}
	b.N *= len(benchPaths)
}

func Benchmark_SetRawInPlace(b *testing.B) {
	data := []byte(benchJSON)
	v1, v2 := []byte(`"1"`), []byte("1")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				SetRawInPlace(data, path, v1)
			case "widget.image.hOffset":
				SetRawInPlace(data, path, v2)
			case "widget.text.onMouseUp":
				SetRawInPlace(data, path, v1)
			}
		}
	}
	b.N *= len(benchPaths)
}

func Benchmark_SJSON_Set(b *testing.B) {
	data := []byte(benchJSON)
	opts := sjson.Options{Optimistic: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				sjson.SetBytesOptions(data, path, "1", &opts)
			case "widget.image.hOffset":
				sjson.SetBytesOptions(data, path, 1, &opts)
			case "widget.text.onMouseUp":
				sjson.SetBytesOptions(data, path, "1", &opts)
			}
		}
	}
	b.N *= len(benchPaths)
}

func Benchmark_SJSON_SetInPlace(b *testing.B) {
	data := []byte(benchJSON)
	opts := sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				sjson.SetBytesOptions(data, path, "1", &opts)
			case "widget.image.hOffset":
				sjson.SetBytesOptions(data, path, 1, &opts)
			case "widget.text.onMouseUp":
				sjson.SetBytesOptions(data, path, "1", &opts)
			}
		}
	}
	b.N *= len(benchPaths)
}

func Benchmark_SJSON_SetRawInPlace(b *testing.B) {
	data := []byte(benchJSON)
	opts := sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	}
	v1, v2 := []byte(`"1"`), []byte("1")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				sjson.SetRawBytesOptions(data, path, v1, &opts)
			case "widget.image.hOffset":
				sjson.SetRawBytesOptions(data, path, v2, &opts)
			case "widget.text.onMouseUp":
				sjson.SetRawBytesOptions(data, path, v1, &opts)
			}
		}
	}
	b.N *= len(benchPaths)
}

func Benchmark_SJSON_SetNoFlag(b *testing.B) {
	data := []byte(benchJSON)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range benchPaths {
			switch path {
			case "widget.window.name":
				sjson.SetBytes(data, path, "1")
			case "widget.image.hOffset":
				sjson.SetBytes(data, path, 1)
			case "widget.text.onMouseUp":
				sjson.SetBytes(data, path, "1")
			}
		}
	}
	b.N *= len(benchPaths)
}
