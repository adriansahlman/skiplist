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
	setupUnsorted := func(s *SkipListFullTestSuite) {
		s.skipList = skiplist.New[int, int](skiplist.WithSeed(seed))
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
	suite.Run(t, &SkipListFullTestSuite{setup: setupUnsorted})
}

func TestSkipListSorted(t *testing.T) {
	setupSorted := func(s *SkipListFullTestSuite) {
		s.skipList = skiplist.New[int, int](skiplist.WithSeed(seed))
		s.allElementsSorted = make([]int, 1<<16)
		for i := range s.allElementsSorted {
			s.allElementsSorted[i] = 2 * i
			s.skipList.Set(s.allElementsSorted[i], s.allElementsSorted[i])
		}
	}
	suite.Run(t, &SkipListFullTestSuite{setup: setupSorted})
}

func TestSkipListReversed(t *testing.T) {
	setupReversed := func(s *SkipListFullTestSuite) {
		s.skipList = skiplist.New[int, int](skiplist.WithSeed(seed))
		s.allElementsSorted = make([]int, 1<<16)
		for i := range s.allElementsSorted {
			s.allElementsSorted[i] = 2 * i
		}
		for i := len(s.allElementsSorted) - 1; i >= 0; i-- {
			s.skipList.Set(s.allElementsSorted[i], s.allElementsSorted[i])
		}
	}
	suite.Run(t, &SkipListFullTestSuite{setup: setupReversed})
}

func BenchmarkSkipList(b *testing.B) {
	rng := rand.New(rand.NewSource(seed))
	for _, shift := range []int{4, 8, 12, 18, 20} {
		elemCount := 1 << shift
		b.Run(fmt.Sprintf("%d_Elements", elemCount), func(b *testing.B) {
			l := skiplist.New[int, int](skiplist.WithSeed(seed))
			elems := make([]int, elemCount)
			for i := 0; i < elemCount; i++ {
				l.Set(i*2, i*2)
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
					l.Get(elemsShifted[i%elemCount])
				}
			})
			b.Run("Before", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					l.Before(elemsShifted[i%elemCount])
				}
			})
			b.Run("After", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					l.After(elemsShifted[i%elemCount])
				}
			})
			b.Run("AtOrBefore", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					l.AtOrBefore(elemsShifted[i%elemCount])
				}
			})
			b.Run("AtOrAfter", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					l.AtOrAfter(elemsShifted[i%elemCount])
				}
			})
			var kv int
			b.Run("RemoveAndSet", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					kv = elems[i%elemCount]
					l.Remove(kv)
					l.Set(kv, kv)
				}
			})
			b.Run("Set_(ExistingElement)", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					kv = elems[i%elemCount]
					l.Set(kv, kv)
				}
			})
			b.Run("Set_(NewElement)", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					kv = elems[i%elemCount] + 1
					l.Set(kv, kv)
					l.RemoveFirst()
				}
			})
		})
	}
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

func (s *SkipListFullTestSuite) TestContains() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyContains := s.skipList.Contains(s.allElementsSorted[i])
		beforeKeyContains := s.skipList.Contains(s.allElementsSorted[i] - 1)
		afterKeyContains := s.skipList.Contains(s.allElementsSorted[i] + 1)
		require.True(t, atKeyContains)
		require.False(t, beforeKeyContains)
		require.False(t, afterKeyContains)
	}
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
	elem := s.skipList.First()
	if len(deleted) < len(s.allElementsSorted) {
		require.NotNil(t, elem)
	}
	for _, v := range s.allElementsSorted {
		if deleted[v] {
			continue
		}
		require.NotNil(t, elem)
		require.Equal(t, v, elem.Value())
		elem = elem.Next()
	}
	elem = s.skipList.Last()
	if len(deleted) < len(s.allElementsSorted) {
		require.NotNil(t, elem)
	}
	var v int
	for i := range s.allElementsSorted {
		v = s.allElementsSorted[len(s.allElementsSorted)-1-i]
		if deleted[v] {
			continue
		}
		require.NotNil(t, elem)
		require.Equal(t, v, elem.Value())
		elem = elem.Prev()
	}
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
	elem := s.skipList.First()
	if len(deleted) < len(s.allElementsSorted) {
		require.NotNil(t, elem)
	}
	for _, v := range s.allElementsSorted {
		if deleted[v] {
			continue
		}
		require.NotNil(t, elem)
		require.Equal(t, v, elem.Value())
		elem = elem.Next()
	}
	elem = s.skipList.Last()
	if len(deleted) < len(s.allElementsSorted) {
		require.NotNil(t, elem)
	}
	var v int
	for i := range s.allElementsSorted {
		v = s.allElementsSorted[len(s.allElementsSorted)-1-i]
		if deleted[v] {
			continue
		}
		require.NotNil(t, elem)
		require.Equal(t, v, elem.Value())
		elem = elem.Prev()
	}
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

func (s *SkipListFullTestSuite) TestBefore() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyElement := s.skipList.Before(s.allElementsSorted[i])
		beforeKeyElement := s.skipList.Before(s.allElementsSorted[i] - 1)
		afterKeyElement := s.skipList.Before(s.allElementsSorted[i] + 1)
		require.NotNil(t, afterKeyElement)
		require.Equal(t, s.allElementsSorted[i], afterKeyElement.Value())
		if i == 0 {
			require.Nil(t, atKeyElement)
			require.Nil(t, beforeKeyElement)
		} else {
			require.NotNil(t, atKeyElement)
			require.Equal(t, s.allElementsSorted[i-1], atKeyElement.Value())
			require.NotNil(t, beforeKeyElement)
			require.Equal(t, s.allElementsSorted[i-1], beforeKeyElement.Value())
		}
	}
}

func (s *SkipListFullTestSuite) TestAfter() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyElement := s.skipList.After(s.allElementsSorted[i])
		beforeKeyElement := s.skipList.After(s.allElementsSorted[i] - 1)
		afterKeyElement := s.skipList.After(s.allElementsSorted[i] + 1)
		require.NotNil(t, beforeKeyElement)
		require.Equal(t, s.allElementsSorted[i], beforeKeyElement.Value())
		if i == len(s.allElementsSorted)-1 {
			require.Nil(t, atKeyElement)
			require.Nil(t, afterKeyElement)
		} else {
			require.NotNil(t, atKeyElement)
			require.Equal(t, s.allElementsSorted[i+1], atKeyElement.Value())
			require.NotNil(t, afterKeyElement)
			require.Equal(t, s.allElementsSorted[i+1], afterKeyElement.Value())
		}
	}
}

func (s *SkipListFullTestSuite) TestAtOrBefore() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyElement := s.skipList.AtOrBefore(s.allElementsSorted[i])
		beforeKeyElement := s.skipList.AtOrBefore(s.allElementsSorted[i] - 1)
		afterKeyElement := s.skipList.AtOrBefore(s.allElementsSorted[i] + 1)
		require.NotNil(t, atKeyElement)
		require.Equal(t, s.allElementsSorted[i], atKeyElement.Value())
		require.NotNil(t, afterKeyElement)
		require.Equal(t, s.allElementsSorted[i], afterKeyElement.Value())
		if i == 0 {
			require.Nil(t, beforeKeyElement)
		} else {
			require.NotNil(t, beforeKeyElement)
			require.Equal(t, s.allElementsSorted[i-1], beforeKeyElement.Value())
		}
	}
}

func (s *SkipListFullTestSuite) TestAtOrAfter() {
	t := s.T()
	for i := range s.allElementsSorted {
		atKeyElement := s.skipList.AtOrAfter(s.allElementsSorted[i])
		beforeKeyElement := s.skipList.AtOrAfter(s.allElementsSorted[i] - 1)
		afterKeyElement := s.skipList.AtOrAfter(s.allElementsSorted[i] + 1)
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
	for i := 0; i < 1<<20; i++ {
		sl.Set(i, struct{}{})
	}

	// Get the element at key 100.
	// Requires an exact match of
	// the key.
	elem := sl.Get(100)
	if elem == nil {
		panic("element should exist")
	}

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
