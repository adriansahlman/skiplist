# Go Generic Skip List

A [skip list](https://en.wikipedia.org/wiki/Skip_list) implemented using generics in Go.

The skip list can be searched for the closest value at/after the search value with an average complexity of O(log(n))

Backwards compatibility may break between minor version updates until v1.0.0 is reached.

## Usage

### Example
```go
package main

import "github.com/adriansahlman/skiplist"

func main() {
	sl := skiplist.NewOrderedMap[int, struct{}]()
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
sl := skiplist.NewOrderedMap[int, string](skiplist.WithHashmap())
```
This reduces the complexity for fetching values for existing keys, as well as setting new values for existing keys to O(1).

### Threadsafety
The skip list is not threadsafe, make sure to use a RW mutex when reading and writing in different simultaneous go routines.

## Benchmarks
Macbook Air M2
```
WithoutHashmap/Length=16/Get                   30.85 ns/op
WithoutHashmap/Length=16/Search                29.35 ns/op
WithoutHashmap/Length=16/Remove->Set           184.8 ns/op
WithoutHashmap/Length=16/Set_(Replace)         124.9 ns/op
WithoutHashmap/Length=16/Set_(New)             103.1 ns/op
WithoutHashmap/Length=256/Get                  45.22 ns/op
WithoutHashmap/Length=256/Search               39.35 ns/op
WithoutHashmap/Length=256/Remove->Set          225.4 ns/op
WithoutHashmap/Length=256/Set_(Replace)        148.7 ns/op
WithoutHashmap/Length=256/Set_(New)            102.5 ns/op
WithoutHashmap/Length=16384/Get                129.2 ns/op
WithoutHashmap/Length=16384/Search             127.4 ns/op
WithoutHashmap/Length=16384/Remove->Set        348.4 ns/op
WithoutHashmap/Length=16384/Set_(Replace)      234.7 ns/op
WithoutHashmap/Length=16384/Set_(New)          152.2 ns/op
WithoutHashmap/Length=262144/Get               292.7 ns/op
WithoutHashmap/Length=262144/Search            272.7 ns/op
WithoutHashmap/Length=262144/Remove->Set       579.1 ns/op
WithoutHashmap/Length=262144/Set_(Replace)     446.5 ns/op
WithoutHashmap/Length=262144/Set_(New)         424.3 ns/op
WithoutHashmap/Length=1048576/Get              506.6 ns/op
WithoutHashmap/Length=1048576/Search           506.0 ns/op
WithoutHashmap/Length=1048576/Remove->Set      843.9 ns/op
WithoutHashmap/Length=1048576/Set_(Replace)    696.6 ns/op
WithoutHashmap/Length=1048576/Set_(New)        704.5 ns/op
WithoutHashmap/Length=16777216/Get             1286 ns/op
WithoutHashmap/Length=16777216/Search          1321 ns/op
WithoutHashmap/Length=16777216/Remove->Set     1533 ns/op
WithoutHashmap/Length=16777216/Set_(Replace)   1541 ns/op
WithoutHashmap/Length=16777216/Set_(New)       1564 ns/op
WithHashmap/Length=16/Get                      9.703 ns/op
WithHashmap/Length=16/Search                   29.68 ns/op
WithHashmap/Length=16/Remove->Set              241.4 ns/op
WithHashmap/Length=16/Set_(Replace)            5.842 ns/op
WithHashmap/Length=16/Set_(New)                119.8 ns/op
WithHashmap/Length=256/Get                     11.23 ns/op
WithHashmap/Length=256/Search                  34.34 ns/op
WithHashmap/Length=256/Remove->Set             305.2 ns/op
WithHashmap/Length=256/Set_(Replace)           6.608 ns/op
WithHashmap/Length=256/Set_(New)               117.2 ns/op
WithHashmap/Length=16384/Get                   14.63 ns/op
WithHashmap/Length=16384/Search                99.21 ns/op
WithHashmap/Length=16384/Remove->Set           457.9 ns/op
WithHashmap/Length=16384/Set_(Replace)         25.17 ns/op
WithHashmap/Length=16384/Set_(New)             183.8 ns/op
WithHashmap/Length=262144/Get                  35.09 ns/op
WithHashmap/Length=262144/Search               264.3 ns/op
WithHashmap/Length=262144/Remove->Set          752.2 ns/op
WithHashmap/Length=262144/Set_(Replace)        51.97 ns/op
WithHashmap/Length=262144/Set_(New)            626.8 ns/op
WithHashmap/Length=1048576/Get                 54.42 ns/op
WithHashmap/Length=1048576/Search              407.8 ns/op
WithHashmap/Length=1048576/Remove->Set         1014 ns/op
WithHashmap/Length=1048576/Set_(Replace)       62.75 ns/op
WithHashmap/Length=1048576/Set_(New)           983.7 ns/op
WithHashmap/Length=16777216/Get                54.27 ns/op
WithHashmap/Length=16777216/Search             999.2 ns/op
WithHashmap/Length=16777216/Remove->Set        1793 ns/op
WithHashmap/Length=16777216/Set_(Replace)      66.08 ns/op
WithHashmap/Length=16777216/Set_(New)          1961 ns/op
```

## License
[MIT](./LICENSE)
