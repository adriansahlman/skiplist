# Go Generic Skip List

A [skip list](https://en.wikipedia.org/wiki/Skip_list) implemented using generics in Go.

This implementation includes additional functionality for finding the closest matching key instead of requiring an exact match.

Backwards compatibility may break between minor version updates until v1.0.0 is reached.

## Usage

### Example
```go
package main

import "github.com/adriansahlman/skiplist"

func main() {
	sl := skiplist.New[int, struct{}]()
	// Fill with even numbers
	for i := 0; i < 1<<20; i++ {
		sl.Set(2*i, struct{}{})
	}

	// Requires an exact match of
	// the key.
	node := sl.Get(100)
	if node == nil {
		panic("node should exist")
	}
	node = sl.Get(101)
	if node != nil {
		panic("node should not exist")
	}

	node = sl.Remove(100)
	if node == nil {
		panic("node should have existed and returned when removed")
	}
	node = sl.Get(100)
	if node != nil {
		panic("node should not exist after being removed")
	}

	node = sl.First()
	if node.Key() != 0 {
		panic("not the first node")
	}
	node = sl.Last()
	if node.Key() != (1<<20-1)*2 {
		panic("not the last node")
	}

	// Get the first node with a key value
	// at or above 101 when traversing
	// the list in ascending order.
	node = sl.Search(101)
	if node.Key() != 102 {
		panic("key != 102")
	}

	// iterate forward through the list
	for node = sl.Get(0); node != nil; node = node.Next() {
		// do something
	}

	// iterate backward through the list
	for node = sl.Get(1000); node != nil; node = node.Prev() {
		// do something
	}
}
```

### Hashmap

If often fetching and writing to keys that already exist in the list it might be a good idea to enable a hashmap through the option `WithHashmap()`.
```go
sl := skiplist.New[int, string](skiplist.WithHashmap())
```
This reduces the complexity for fetching values for existing keys, as well as setting new values for existing keys to O(1).

### Threadsafety
The skip list is not threadsafe, make sure to use an RW mutex when reading and writing in different simultaneous go routines.

## Benchmarks
Macbook Air M2
```
BenchmarkAll/WithoutHashmap/Length=16/Get                         28.34 ns/op
BenchmarkAll/WithoutHashmap/Length=16/Search                      30.54 ns/op
BenchmarkAll/WithoutHashmap/Length=16/Remove->Set                 221.0 ns/op
BenchmarkAll/WithoutHashmap/Length=16/Set_(Replace)               250.3 ns/op
BenchmarkAll/WithoutHashmap/Length=16/Set_(New)                   189.7 ns/op
BenchmarkAll/WithoutHashmap/Length=256/Get                        40.75 ns/op
BenchmarkAll/WithoutHashmap/Length=256/Search                     39.49 ns/op
BenchmarkAll/WithoutHashmap/Length=256/Remove->Set                247.8 ns/op
BenchmarkAll/WithoutHashmap/Length=256/Set_(Replace)              294.4 ns/op
BenchmarkAll/WithoutHashmap/Length=256/Set_(New)                  196.8 ns/op
BenchmarkAll/WithoutHashmap/Length=16384/Get                      129.7 ns/op
BenchmarkAll/WithoutHashmap/Length=16384/Search                   127.5 ns/op
BenchmarkAll/WithoutHashmap/Length=16384/Remove->Set              389.7 ns/op
BenchmarkAll/WithoutHashmap/Length=16384/Set_(Replace)            282.8 ns/op
BenchmarkAll/WithoutHashmap/Length=16384/Set_(New)                204.0 ns/op
BenchmarkAll/WithoutHashmap/Length=262144/Get                     435.8 ns/op
BenchmarkAll/WithoutHashmap/Length=262144/Search                  415.4 ns/op
BenchmarkAll/WithoutHashmap/Length=262144/Remove->Set             965.2 ns/op
BenchmarkAll/WithoutHashmap/Length=262144/Set_(Replace)           703.7 ns/op
BenchmarkAll/WithoutHashmap/Length=262144/Set_(New)               688.7 ns/op
BenchmarkAll/WithoutHashmap/Length=1048576/Get                    689.3 ns/op
BenchmarkAll/WithoutHashmap/Length=1048576/Search                 695.7 ns/op
BenchmarkAll/WithoutHashmap/Length=1048576/Remove->Set            1256 ns/op
BenchmarkAll/WithoutHashmap/Length=1048576/Set_(Replace)          1040 ns/op
BenchmarkAll/WithoutHashmap/Length=1048576/Set_(New)              1048 ns/op
BenchmarkAll/WithoutHashmap/Length=16777216/Get                   1632 ns/op
BenchmarkAll/WithoutHashmap/Length=16777216/Search                1674 ns/op
BenchmarkAll/WithoutHashmap/Length=16777216/Remove->Set           2514 ns/op
BenchmarkAll/WithoutHashmap/Length=16777216/Set_(Replace)         2367 ns/op
BenchmarkAll/WithoutHashmap/Length=16777216/Set_(New)             2218 ns/op
BenchmarkAll/WithHashmap/Length=16/Get                            9.722 ns/op
BenchmarkAll/WithHashmap/Length=16/Search                         28.49 ns/op
BenchmarkAll/WithHashmap/Length=16/Remove->Set                    62.81 ns/op
BenchmarkAll/WithHashmap/Length=16/Set_(Replace)                  5.454 ns/op
BenchmarkAll/WithHashmap/Length=16/Set_(New)                      192.7 ns/op
BenchmarkAll/WithHashmap/Length=256/Get                           11.91 ns/op
BenchmarkAll/WithHashmap/Length=256/Search                        35.10 ns/op
BenchmarkAll/WithHashmap/Length=256/Remove->Set                   69.55 ns/op
BenchmarkAll/WithHashmap/Length=256/Set_(Replace)                 6.616 ns/op
BenchmarkAll/WithHashmap/Length=256/Set_(New)                     208.8 ns/op
BenchmarkAll/WithHashmap/Length=16384/Get                         15.99 ns/op
BenchmarkAll/WithHashmap/Length=16384/Search                      104.0 ns/op
BenchmarkAll/WithHashmap/Length=16384/Remove->Set                 92.06 ns/op
BenchmarkAll/WithHashmap/Length=16384/Set_(Replace)               26.53 ns/op
BenchmarkAll/WithHashmap/Length=16384/Set_(New)                   171.5 ns/op
BenchmarkAll/WithHashmap/Length=262144/Get                        45.37 ns/op
BenchmarkAll/WithHashmap/Length=262144/Search                     412.7 ns/op
BenchmarkAll/WithHashmap/Length=262144/Remove->Set                637.8 ns/op
BenchmarkAll/WithHashmap/Length=262144/Set_(Replace)              51.84 ns/op
BenchmarkAll/WithHashmap/Length=262144/Set_(New)                  770.0 ns/op
BenchmarkAll/WithHashmap/Length=1048576/Get                       59.02 ns/op
BenchmarkAll/WithHashmap/Length=1048576/Search                    627.8 ns/op
BenchmarkAll/WithHashmap/Length=1048576/Remove->Set               959.0 ns/op
BenchmarkAll/WithHashmap/Length=1048576/Set_(Replace)             62.72 ns/op
BenchmarkAll/WithHashmap/Length=1048576/Set_(New)                 1298 ns/op
BenchmarkAll/WithHashmap/Length=16777216/Get                      68.17 ns/op
BenchmarkAll/WithHashmap/Length=16777216/Search                   1345 ns/op
BenchmarkAll/WithHashmap/Length=16777216/Remove->Set              2326 ns/op
BenchmarkAll/WithHashmap/Length=16777216/Set_(Replace)            75.58 ns/op
BenchmarkAll/WithHashmap/Length=16777216/Set_(New)                2423 ns/op
```

## License
[MIT](./LICENSE)
