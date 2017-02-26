mjson
-----

[![GoDoc](https://godoc.org/github.com/lukechampine/mjson?status.svg)](https://godoc.org/github.com/lukechampine/mjson)
[![Go Report Card](http://goreportcard.com/badge/github.com/lukechampine/mjson)](https://goreportcard.com/report/github.com/lukechampine/mjson)

```
go get github.com/lukechampine/mjson
```

`mjson` sets values in JSON super fast. It is comparable to [SJSON](https://github.com/tidwall/sjson), but
with some key differences. It was created to support the [`jj`](https://github.com/lukechampine/jj) transaction journal.

Unlike SJSON, `mjson` does not support deletion, creating nested objects,
escaping the `.` character in paths, or the special `-1` index. However, it
does support appending to `null` as though it were `[]`, and does not require
the special `:` syntax for integer object keys. Appending to an array is still
possible as well, using the length of the array as an index. This is safer
than using `-1` because setting an explicit index is an idempotent operation,
whereas `-1` is context-sensitive.


## Examples ##

Set an existing object key:
```go
json := mjson.Set(`{"name":{"last":"Anderson"}}`, "name.last", "Smith")
// {"name":{"last":"Smith"}}
```

Add a new object key:
```go
json := mjson.Set(`{"name":{"last":"Anderson"}}`, "name.first", "Sara")
// {"name":{"first":"Sara","last":"Anderson"}}
```

Set an existing array value:
```go
json := mjson.Set(`{"friends":["Andy","Carol"]}`, "friends.1", "Sara")
// {"friends":["Andy","Sara"]
```

Append a new array value:
```go
json := mjson.Set(`{"friends":["Andy","Carol"]}`, "friends.2", "Sara")
// {"friends":["Andy","Carol","Sara"]
```


## Benchmarks ##

`mjson` runs a teeny bit faster than SJSON:

```
Benchmark_Set-4                   	 3000000	       556 ns/op	     405 B/op	       2 allocs/op
Benchmark_SetInPlace-4            	 3000000	       448 ns/op	      21 B/op	       2 allocs/op
Benchmark_SetRawInPlace-4         	 3000000	       346 ns/op	       0 B/op	       0 allocs/op
Benchmark_SJSON_Set-4             	 3000000	       694 ns/op	     576 B/op	       2 allocs/op
Benchmark_SJSON_SetInPlace-4      	 3000000	       564 ns/op	      74 B/op	       1 allocs/op
Benchmark_SJSON_SetRawInPlace-4   	 3000000	       426 ns/op	       0 B/op	       0 allocs/op
```

However, to get this speed with SJSON you must set the `Optimistic` flag.
Otherwise its performance drops considerably:

```
Benchmark_SJSON_SetNoFlag-4        	 3000000	      1861 ns/op	    1178 B/op	       5 allocs/op
```

You should use whichever API you prefer, though.