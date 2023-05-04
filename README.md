# Go Generic Skip List

A [skip list](https://en.wikipedia.org/wiki/Skip_list) implemented using generics in Go.

This implementation includes additional functionality for finding the closest matching key instead of requiring an exact match.

## Usage
```go
package main

import "github.com/adriansahlman/skiplist"

func main() {
    sl := skiplist.New[int, struct{}]()
    for i := 0; i < 1<<20; i++ {
        sl.Set(i, struct{}{})
    }

    // Get the element at key 100.
    // Requires an exact match of
    // the key.
    elem := sl.Get(100)

    // Get the element with a key value
    // lower than 100 (prioritizing the highest
    // valued key), in this case 99.
    elem = sl.Before(100)
    if elem.Key() != 99 {
        panic("key != 99")
    }

    // Get the element with a key value
    // at or lower lower than 100
    // (prioritizing the highest
    // valued key), in this case 100.
    elem = sl.AtOrBefore(100)
    if elem.Key() != 100 {
        panic("key != 100")
    }

    elem = sl.After(100)
    if elem.Key() != 101 {
        panic("key != 101")
    }

    elem = sl.AtOrAfter(100)
    if elem.Key() != 100 {
        panic("key != 100")
    }

    // iterate forward through the list
    for elem = sl.Get(100); elem != nil; elem = elem.Next() {
        // do something
    }

    // iterate backward through the list
    for elem = sl.Get(100); elem != nil; elem = elem.Prev() {
        // do something
    }
}
```

## License
[MIT](./LICENSE)
