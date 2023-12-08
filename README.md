# Go Generic Skip List

A [skip list](https://en.wikipedia.org/wiki/Skip_list) implemented using generics in Go.

Adding, removing and searching for values should have an average complexity of O(log(n)).

The implementation is not threadsafe.

## Usage

```go
package main

import "github.com/adriansahlman/skiplist"

func main() {
	// var list *skiplist.SkipList[int]
	list := skiplist.New(func(a, b int) bool { return a < b })

	// Add some values
	for i := 0; i < 16; i++ {
		list.Add(i)
	}

	// Iterate over values in ascending order
	for node := list.First(); node != nil; node = node.Next() {
		_ = node.Value()
	}

	// Iterate over values in descending order
	for node := list.Last(); node != nil; node = node.Prev() {
		_ = node.Value()
	}

	// Remove a value
	node := list.Remove(3)
	// Check if the value was found and removed
	if node == nil || node.Value() != 3 {
		panic(node)
	}

	// Removing the first value (ascending order)
	// has a complexity of O(1).
	node = list.RemoveFirst()
	if node == nil || node.Value() != 0 {
		panic(node)
	}

	// Find the first node (ascending order) with
	// a value greater than or equal to the given value.
	node = list.Search(2)
	if node == nil || node.Value() != 2 {
		panic(node)
	}
	// As we removed the value 3 we shoud get the next
	// value which in this case is 4.
	node = list.Search(3)
	if node == nil || node.Value() != 4 {
		panic(node)
	}

	// The skiplist can act as a set (duplicate values
	// will not occur in the list) with an option.
	list = skiplist.New(
		func(a, b int) bool { return a < b },
		skiplist.WithReplace(),
	)

	for i := 0; i < 16; i++ {
		list.Add(0)
	}
	if list.Length() != 1 {
		panic(list.Length())
	}

}
```

## License
[MIT](./LICENSE)
