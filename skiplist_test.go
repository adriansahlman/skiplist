package skiplist_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/adriansahlman/skiplist"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const seed = 0

func TestSkipListUnsorted(t *testing.T) {
	var enableHashmap bool
	setupUnsorted := func(s *SkipListFullTestSuite) {
		opts := []skiplist.Option{
			skiplist.WithSeed(seed),
		}
		if enableHashmap {
			opts = append(opts, skiplist.WithHashmap())
		}
		s.skipList = skiplist.New[int, int](opts...)
		data := []int{4, 2, 10, 6, 8, 12, 18, 14, 16}
		for _, kv := range data {
			s.skipList.Set(kv, kv)
		}
		// add twice, should overwrite prev
		for i := range data {
			v := data[len(data)-1-i]
			s.skipList.Set(v, v)
		}
		sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
		s.allElementsSorted = data
	}
	for _, enableHashmap = range []bool{false, true} {
		name := "WithoutHashmap"
		if enableHashmap {
			name = "WithHashmap"
		}
		t.Run(name, func(t *testing.T) {
			suite.Run(t, &SkipListFullTestSuite{setup: setupUnsorted})
		})
	}
}

func TestSkipListSorted(t *testing.T) {
	var enableHashmap bool
	setupSorted := func(s *SkipListFullTestSuite) {
		opts := []skiplist.Option{
			skiplist.WithSeed(seed),
		}
		if enableHashmap {
			opts = append(opts, skiplist.WithHashmap())
		}
		s.skipList = skiplist.New[int, int](opts...)
		s.allElementsSorted = make([]int, 1<<16)
		for i := range s.allElementsSorted {
			s.allElementsSorted[i] = 2 * i
			s.skipList.Set(s.allElementsSorted[i], s.allElementsSorted[i])
		}
	}
	for _, enableHashmap = range []bool{false, true} {
		name := "WithoutHashmap"
		if enableHashmap {
			name = "WithHashmap"
		}
		t.Run(name, func(t *testing.T) {
			suite.Run(t, &SkipListFullTestSuite{setup: setupSorted})
		})
	}
}

func TestSkipListReversed(t *testing.T) {
	var enableHashmap bool
	setupReversed := func(s *SkipListFullTestSuite) {
		opts := []skiplist.Option{
			skiplist.WithSeed(seed),
		}
		if enableHashmap {
			opts = append(opts, skiplist.WithHashmap())
		}
		s.skipList = skiplist.New[int, int](opts...)
		s.allElementsSorted = make([]int, 1<<16)
		for i := range s.allElementsSorted {
			s.allElementsSorted[i] = 2 * i
		}
		for i := len(s.allElementsSorted) - 1; i >= 0; i-- {
			s.skipList.Set(s.allElementsSorted[i], s.allElementsSorted[i])
		}
	}
	for _, enableHashmap = range []bool{false, true} {
		name := "WithoutHashmap"
		if enableHashmap {
			name = "WithHashmap"
		}
		t.Run(name, func(t *testing.T) {
			suite.Run(t, &SkipListFullTestSuite{setup: setupReversed})
		})
	}
}

func BenchmarkAll(b *testing.B) {
	for _, enableHashmap := range []bool{false, true} {
		opts := []skiplist.Option{skiplist.WithSeed(seed)}
		name := "WithoutHashmap"
		if enableHashmap {
			name = "WithHashmap"
			opts = append(opts, skiplist.WithHashmap())
		}
		b.Run(name, func(b *testing.B) {
			for _, shift := range []int{4, 8, 14, 18, 20, 24} {
				n := 1 << shift
				name := fmt.Sprintf("Length=%d", n)
				b.Run(name, func(b *testing.B) {
					benchmarkSkipListFunctions(
						b,
						skiplist.New[int, struct{}](opts...),
						n,
					)
				})
			}
		})
	}
}

func benchmarkSkipListFunctions(
	b *testing.B,
	l *skiplist.SkipList[int, struct{}],
	n int,
) {
	rng := rand.New(rand.NewSource(seed))
	elems := make([]int, n)
	for i := 0; i < n; i++ {
		l.Set(i*2, struct{}{})
		elems[i] = i * 2
	}
	rng.Shuffle(
		len(elems),
		func(i, j int) { elems[i], elems[j] = elems[j], elems[i] },
	)
	elemsShifted := make([]int, len(elems))
	copy(elemsShifted, elems)
	for i := range elemsShifted {
		elemsShifted[i] += rng.Intn(3) - 1
	}
	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l.Get(elemsShifted[i%n])
		}
	})
	b.Run("Search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l.Search(elemsShifted[i%n])
		}
	})
	var kv int
	b.Run("Remove->Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			kv = elems[i%n]
			l.Remove(kv)
			l.Set(kv, struct{}{})
		}
	})
	b.Run("Set_(Replace)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			kv = elems[i%n]
			l.Set(kv, struct{}{})
		}
	})
	b.Run("Set_(New)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			kv = elems[i%n]
			l.Set(kv+1, struct{}{})
			l.RemoveFirst()
		}
	})
}

var _ suite.BeforeTest = (*SkipListFullTestSuite)(nil)

// Each element is used as both key and value
// in skiplist.
// All elements are even.
type SkipListFullTestSuite struct {
	suite.Suite
	skipList          *skiplist.SkipList[int, int]
	allElementsSorted []int
	setup             func(*SkipListFullTestSuite)
}

// BeforeTest implements suite.BeforeTest
func (s *SkipListFullTestSuite) BeforeTest(suiteName string, testName string) {
	require.NotNil(s.T(), s.setup)
	s.setup(s)
}

func (s *SkipListFullTestSuite) TestLength() {
	t := s.T()
	require.Equal(t, len(s.allElementsSorted), s.skipList.Length())
}

func (s *SkipListFullTestSuite) TestRemove() {
	t := s.T()
	deleted := make(map[int]bool, len(s.allElementsSorted))
	for _, i := range []int{0, len(s.allElementsSorted) - 1, len(s.allElementsSorted) / 2} {
		if deleted[s.allElementsSorted[i]] {
			continue
		}
		got := s.skipList.Remove(s.allElementsSorted[i])
		require.NotNil(t, got)
		require.Equal(t, s.allElementsSorted[i], got.Value())
		got = s.skipList.Remove(s.allElementsSorted[i])
		require.Nil(t, got)
		deleted[s.allElementsSorted[i]] = true
		require.Equal(
			t,
			len(s.allElementsSorted)-len(deleted),
			s.skipList.Length(),
		)
	}
	require.Equal(t, len(s.allElementsSorted)-len(deleted), s.skipList.Length())
	testForwardIter := func() {
		node := s.skipList.First()
		if len(deleted) < len(s.allElementsSorted) {
			require.NotNil(t, node)
		}
		for _, v := range s.allElementsSorted {
			if deleted[v] {
				continue
			}
			require.NotNil(t, node)
			require.Equal(t, v, node.Value())
			node = node.Next()
		}
	}
	testBackwardIter := func() {
		node := s.skipList.Last()
		if len(deleted) < len(s.allElementsSorted) {
			require.NotNil(t, node)
		}
		var v int
		for i := range s.allElementsSorted {
			v = s.allElementsSorted[len(s.allElementsSorted)-1-i]
			if deleted[v] {
				continue
			}
			require.NotNil(t, node)
			require.Equal(t, v, node.Value())
			node = node.Prev()
		}
	}
	testForwardIter()
	testBackwardIter()
	for kv := range deleted {
		s.skipList.Set(kv, kv)
		delete(deleted, kv)
	}
	testForwardIter()
	testBackwardIter()
}

func (s *SkipListFullTestSuite) TestRemoveFirst() {
	t := s.T()
	deleted := make(map[int]bool, len(s.allElementsSorted))
	n := 3
	if len(s.allElementsSorted) < n {
		n = len(s.allElementsSorted)
	}
	for i := 0; i < n; i++ {
		deleted[s.skipList.RemoveFirst().Key()] = true
	}
	require.Equal(t, n, len(deleted))
	require.Equal(t, len(s.allElementsSorted)-n, s.skipList.Length())
	testForwardIter := func() {
		node := s.skipList.First()
		if len(deleted) < len(s.allElementsSorted) {
			require.NotNil(t, node)
		}
		for _, v := range s.allElementsSorted {
			if deleted[v] {
				continue
			}
			require.NotNil(t, node)
			require.Equal(t, v, node.Value())
			node = node.Next()
		}
	}
	testBackwardIter := func() {
		node := s.skipList.Last()
		if len(deleted) < len(s.allElementsSorted) {
			require.NotNil(t, node)
		}
		var v int
		for i := range s.allElementsSorted {
			v = s.allElementsSorted[len(s.allElementsSorted)-1-i]
			if deleted[v] {
				continue
			}
			require.NotNil(t, node)
			require.Equal(t, v, node.Value())
			node = node.Prev()
		}
	}
	testForwardIter()
	testBackwardIter()
	for kv := range deleted {
		s.skipList.Set(kv, kv)
		delete(deleted, kv)
	}
	testForwardIter()
	testBackwardIter()
	var v int
	var node *skiplist.Node[int, int]
	for i := range s.allElementsSorted {
		v = s.allElementsSorted[i]
		node = s.skipList.RemoveFirst()
		require.NotNil(t, node)
		require.Equal(t, v, node.Value())
	}
	node = s.skipList.RemoveFirst()
	require.Nil(t, node)
	node = s.skipList.First()
	require.Nil(t, node)
	node = s.skipList.Last()
	require.Nil(t, node)
}

func (s *SkipListFullTestSuite) TestGet() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyElement := s.skipList.Get(s.allElementsSorted[i])
		beforeKeyElement := s.skipList.Get(s.allElementsSorted[i] - 1)
		afterKeyElement := s.skipList.Get(s.allElementsSorted[i] + 1)
		require.NotNil(t, atKeyElement)
		require.Equal(t, s.allElementsSorted[i], atKeyElement.Value())
		require.Nil(t, beforeKeyElement)
		require.Nil(t, afterKeyElement)
	}
}

func (s *SkipListFullTestSuite) TestSearch() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyElement := s.skipList.Search(s.allElementsSorted[i])
		beforeKeyElement := s.skipList.Search(
			s.allElementsSorted[i] - 1,
		)
		afterKeyElement := s.skipList.Search(
			s.allElementsSorted[i] + 1,
		)
		require.NotNil(t, atKeyElement)
		require.Equal(t, s.allElementsSorted[i], atKeyElement.Value())
		require.NotNil(t, beforeKeyElement)
		require.Equal(t, s.allElementsSorted[i], beforeKeyElement.Value())
		if i == len(s.allElementsSorted)-1 {
			require.Nil(t, afterKeyElement)
		} else {
			require.NotNil(t, afterKeyElement)
			require.Equal(t, s.allElementsSorted[i+1], afterKeyElement.Value())
		}
	}
}

func Example() {
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
