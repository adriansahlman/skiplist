package skiplist_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/adriansahlman/skiplist"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

func less[T constraints.Ordered](a, b T) bool { return a < b }
func addAll[T any](t testing.TB, sl *skiplist.SkipList[T], data []T) {
	for i := range data {
		n, _ := sl.Add(data[i])
		require.NotNil(t, n)
	}
}

func requireEqual[T comparable](
	t *testing.T,
	sl *skiplist.SkipList[T],
	sortedData []T,
) {
	t.Run("Length", func(t *testing.T) {
		require.Equal(t, len(sortedData), sl.Length())
	})
	t.Run("First", func(t *testing.T) {
		node := sl.First()
		if len(sortedData) == 0 {
			require.Nil(t, node)
			return
		}
		require.NotNil(t, node)
		require.Equal(t, sortedData[0], node.Value())
	})
	t.Run("Last", func(t *testing.T) {
		node := sl.Last()
		if len(sortedData) == 0 {
			require.Nil(t, node)
			return
		}
		require.NotNil(t, node)
		require.Equal(t, sortedData[len(sortedData)-1], node.Value())
	})
	t.Run("Next", func(t *testing.T) {
		node := sl.First()
		for i := range sortedData {
			require.NotNil(t, node)
			require.Equal(t, sortedData[i], node.Value())
			node = node.Next()
		}
	})
	t.Run("Prev", func(t *testing.T) {
		node := sl.Last()
		for i := range sortedData {
			require.NotNil(t, node)
			require.Equal(t, sortedData[len(sortedData)-1-i], node.Value())
			node = node.Prev()
		}
	})
}

func TestAdd(t *testing.T) {
	const numElem = 1 << 16
	sortedData := [numElem]int{}
	for i := 0; i < numElem; i++ {
		sortedData[i] = i
	}
	for _, name := range [...]string{"Ascending", "Descending", "RandomOrder"} {
		var testData []int
		switch name {
		case "Ascending":
			testData = sortedData[:]
		case "Descending":
			testData = make([]int, len(sortedData))
			for i := range sortedData {
				testData[len(sortedData)-1-i] = sortedData[i]
			}
		case "RandomOrder":
			testData = make([]int, len(sortedData))
			copy(testData, sortedData[:])
			rand.Shuffle(
				len(testData),
				func(i, j int) { testData[i], testData[j] = testData[j], testData[i] },
			)
		default:
			panic(name)
		}
		t.Run(name, func(t *testing.T) {
			sl := skiplist.New(less[int])
			addAll(t, sl, testData)
			requireEqual(t, sl, sortedData[:])
			addAll(t, sl, testData)
			expectedData := make([]int, 2*len(sortedData))
			for i := range expectedData {
				expectedData[i] = sortedData[i/2]
			}
			requireEqual(t, sl, expectedData)
			t.Run("WithReplace", func(t *testing.T) {
				sl := skiplist.New(less[int], skiplist.WithReplace())
				for i := range testData {
					n, replaced := sl.Add(testData[i])
					require.NotNil(t, n)
					require.Nil(t, replaced)
				}
				requireEqual(t, sl, sortedData[:])
				for i := range testData {
					n, replaced := sl.Add(testData[i])
					require.NotNil(t, n)
					require.NotNil(t, replaced)
				}
				requireEqual(t, sl, sortedData[:])
			})
		})
	}
	t.Run("Complexity", func(t *testing.T) {
		expectedComplexity := math.Log2(float64(len(sortedData)))
		// Complexity limit of 3*Log(n)
		maxComplexity := 3 * expectedComplexity
		counter := new(int)
		lessWithCount := func(a, b int) bool {
			(*counter)++
			return a < b
		}
		sl := skiplist.New(lessWithCount)
		addAll(t, sl, sortedData[:])
		totalCount := 0
		for i := range sortedData {
			*counter = 0
			node, _ := sl.Add(sortedData[i])
			require.NotNil(t, node)
			totalCount += *counter
			require.NotNil(t, node.RemoveFrom(sl))
		}
		avgComplexity := float64(totalCount) / float64(len(sortedData))
		if avgComplexity > maxComplexity {
			t.Errorf(
				"expected a complexity of %.2f, got %.2f",
				expectedComplexity,
				avgComplexity,
			)
		}
		t.Run("WithReplace", func(t *testing.T) {
			// Additional comparisons are made to check
			// for equality, therefore we must increase
			// the complexity limit to 5*Log(n).
			maxComplexity := 5 * expectedComplexity
			sl := skiplist.New(lessWithCount, skiplist.WithReplace())
			addAll(t, sl, sortedData[:])
			totalCount := 0
			for i := range sortedData {
				*counter = 0
				node, replaced := sl.Add(sortedData[i])
				require.NotNil(t, node)
				require.NotNil(t, replaced)
				totalCount += *counter
			}
			avgComplexity := float64(totalCount) / float64(len(sortedData))
			if avgComplexity > maxComplexity {
				t.Errorf(
					"expected a complexity of %.2f, got %.2f",
					expectedComplexity,
					avgComplexity,
				)
			}
		})
	})
}

func TestRemoveFrom(t *testing.T) {
	const numElem = 1 << 16
	sortedData := [numElem]int{}
	for i := 0; i < numElem; i++ {
		sortedData[i] = i
	}
	sl := skiplist.New(less[int])
	addAll(t, sl, sortedData[:])
	for i := range sortedData {
		node := sl.First()
		require.NotNil(t, node)
		require.Equal(t, sortedData[i], node.Value())
		require.NotNil(t, node.RemoveFrom(sl))
		require.Equal(t, len(sortedData)-1-i, sl.Length())
	}
	require.Nil(t, sl.First())
	require.Nil(t, sl.Last())
	addAll(t, sl, sortedData[:])
	for i := range sortedData {
		node := sl.Last()
		require.NotNil(t, node)
		require.Equal(t, sortedData[len(sortedData)-1-i], node.Value())
		require.NotNil(t, node.RemoveFrom(sl))
		require.Equal(t, len(sortedData)-1-i, sl.Length())
	}
	t.Run("Duplicates", func(t *testing.T) {
		type kv struct{ key, value int }
		lessKey := func(a, b kv) bool {
			return a.key < b.key
		}
		sortedData := make([]kv, 512)
		for i := range sortedData {
			sortedData[i].value = i
		}
		sl := skiplist.New(lessKey)
		addAll(t, sl, sortedData)
		node := sl.First()
		require.NotNil(t, node)
		// go to middle of list
		for i := 0; i < 256; i++ {
			node = node.Next()
			require.NotNil(t, node)
		}
		removedTuple := node.Value()
		node = node.RemoveFrom(sl)
		require.NotNil(t, node)
		require.Equal(t, removedTuple, node.Value())
		require.Nil(t, node.RemoveFrom(sl))
		require.Equal(t, len(sortedData)-1, sl.Length())
		node = sl.First()
		for i := 0; i < 512-1; i++ {
			require.NotNil(t, node)
			require.NotEqual(t, removedTuple, node.Value())
			node = node.Next()
		}
	})
}

func TestRemove(t *testing.T) {
	const numElem = 1 << 16
	sortedData := [numElem]int{}
	for i := 0; i < numElem; i++ {
		sortedData[i] = i
	}
	t.Run("Ascending", func(t *testing.T) {
		sl := skiplist.New(less[int])
		addAll(t, sl, sortedData[:])
		for i := range sortedData {
			require.NotNil(t, sl.First())
			require.NotNil(t, sl.Last())
			require.Equal(t, sortedData[i], sl.First().Value())
			node := sl.Remove(sortedData[i])
			require.NotNil(t, node)
			require.Equal(t, sortedData[i], node.Value())
			require.Equal(t, len(sortedData)-i-1, sl.Length())
		}
	})
	t.Run("Descending", func(t *testing.T) {
		sl := skiplist.New(less[int])
		addAll(t, sl, sortedData[:])
		for i := range sortedData {
			require.NotNil(t, sl.First())
			require.NotNil(t, sl.Last())
			require.Equal(t, sortedData[len(sortedData)-1-i], sl.Last().Value())
			node := sl.Remove(sortedData[len(sortedData)-1-i])
			require.NotNil(t, node)
			require.Equal(t, sortedData[len(sortedData)-1-i], node.Value())
			require.Equal(t, len(sortedData)-i-1, sl.Length())
		}
	})
	t.Run("Random", func(t *testing.T) {
		order := make([]int, len(sortedData))
		for i := range order {
			order[i] = i
		}
		rand.Shuffle(
			len(order),
			func(i, j int) { order[i], order[j] = order[j], order[i] },
		)
		sl := skiplist.New(less[int])
		addAll(t, sl, sortedData[:])
		for i := range order {
			value := sortedData[order[i]]
			require.NotNil(t, sl.First())
			require.NotNil(t, sl.Last())
			node := sl.Remove(value)
			require.NotNil(t, node)
			require.Equal(t, value, node.Value())
			require.Equal(t, len(sortedData)-i-1, sl.Length())
		}
	})
}

func TestRemoveFirst(t *testing.T) {
	const numElem = 1 << 16
	sortedData := [numElem]int{}
	for i := 0; i < numElem; i++ {
		sortedData[i] = i
	}
	sl := skiplist.New(less[int])
	addAll(t, sl, sortedData[:])
	for i := range sortedData {
		require.NotNil(t, sl.First())
		require.NotNil(t, sl.Last())
		require.Equal(t, sortedData[i], sl.First().Value())
		node := sl.RemoveFirst()
		require.NotNil(t, node)
		require.Equal(t, sortedData[i], node.Value())
		require.Equal(t, len(sortedData)-i-1, sl.Length())
	}
}

func TestSearch(t *testing.T) {
	const numElem = 1 << 16
	sortedData := [numElem]float64{}
	for i := 0; i < numElem; i++ {
		sortedData[i] = float64(i)
	}
	sl := skiplist.New(less[float64])
	addAll(t, sl, sortedData[:])
	var node *skiplist.Node[float64]
	for i := range sortedData {
		node = sl.Search(sortedData[i])
		require.NotNil(t, node)
		require.Equal(t, sortedData[i], node.Value())
		node = sl.Search(sortedData[i] - 0.5)
		require.NotNil(t, node)
		require.Equal(t, sortedData[i], node.Value())
	}
	node = sl.Search(sortedData[len(sortedData)-1] + 10)
	require.Nil(t, node)
}

func ExampleSkipList() {
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
