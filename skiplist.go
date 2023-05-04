// Package skiplist implements the skip list datastructure.
package skiplist

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/exp/constraints"
)

const (
	// default static values for options
	DefaultMaxLevel    = 32
	DefaultProbability = 0.5
	DefaultThreadsafe  = false
	DefaultHashmap     = true
)

type Element[K, V any] struct {
	key   K
	value V

	// The skip-lanes. The first lane
	// (which is the element directly succeeding
	// this element) is always present (len(lanes) >= 1).
	// The value of a lane may be nil if
	// there is no succeeding element in
	// that lane.
	lanes []*Element[K, V]
	// The element directly preceeding this element
	// in the list
	prev *Element[K, V]
}

// Get the element directly succeeding this element.
// Returns nil if no succeeding element exists.
// Not threadsafe.
func (e *Element[K, V]) Next() (next *Element[K, V]) {
	if e != nil {
		next = e.lanes[0]
	}
	return
}

// Get the element directly preceeding this element.
// Returns nil if no preceeding element exists.
// Not threadsafe.
func (e *Element[K, V]) Prev() (prev *Element[K, V]) {
	if e != nil {
		prev = e.prev
	}
	return
}

// Get the key of this element.
func (e *Element[K, V]) Key() (key K) {
	if e != nil {
		key = e.key
	}
	return
}

// Get the value of this element.
func (e *Element[K, V]) Value() (value V) {
	if e != nil {
		value = e.value
	}
	return
}

// Implements a skip-list (doubly linked, singly linked skip lanes).
type SkipList[K constraints.Ordered, V any] struct {
	// fast lookup of exact key
	elements map[K]*Element[K, V]
	// skip-lanes
	header []*Element[K, V]
	// the last element in the list
	last *Element[K, V]
	// number of elements
	length int
	// the maximum number of levels for
	// elements in the list
	maxLevel int
	// probability of each level increase
	// for new elements
	prob float64
	// random number generator used throughout
	// the list operations
	rng *rand.Rand
	// optional threadsafety
	rwMutex *sync.RWMutex
}

// Create a new skiplist. Panics if invalid
// options are given (for example, negative probability).
// Not thread safe by default, use `WithThreadSafe()`
// option to add thread safety.
func New[K constraints.Ordered, V any](
	opts ...SkipListOption,
) *SkipList[K, V] {
	o := skipListOptions{
		maxLevel:   DefaultMaxLevel,
		prob:       DefaultProbability,
		seed:       time.Now().UnixNano(),
		threadSafe: DefaultThreadsafe,
		hashmap:    DefaultHashmap,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	if err := o.validate(); err != nil {
		panic(err)
	}
	var lock *sync.RWMutex
	if o.threadSafe {
		lock = &sync.RWMutex{}
	}
	var elements map[K]*Element[K, V]
	if o.hashmap {
		elements = make(map[K]*Element[K, V])
	}
	return &SkipList[K, V]{
		elements: elements,
		header:   make([]*Element[K, V], o.maxLevel),
		maxLevel: o.maxLevel,
		prob:     o.prob,
		rng:      rand.New(rand.NewSource(o.seed)),
		rwMutex:  lock,
	}
}

// Returns the number of key-value pairs in the list
func (l *SkipList[K, V]) Length() int {
	if l.rwMutex != nil {
		l.rwMutex.RLock()
		defer l.rwMutex.RUnlock()
	}
	return l.length
}

// Returns the first element of the list or nil
// if the list is empty
func (l *SkipList[K, V]) First() *Element[K, V] {
	return l.header[0]
}

// Returns the last element of the list or nil
// if the list is empty
func (l *SkipList[K, V]) Last() *Element[K, V] {
	return l.last
}

// Set the value for a key. If the key already exists, the
// levels (skips) are kept and the old value is replaced.
// Returns the element of the key-value pair.
// Complexity: O(1) with hashmap and the key already exists, else O(log(n))
func (l *SkipList[K, V]) Set(key K, value V) *Element[K, V] {
	if l.rwMutex != nil {
		l.rwMutex.Lock()
		defer l.rwMutex.Unlock()
	}
	elem := l.get(key)
	if elem != nil {
		elem.key = key
		elem.value = value
		return elem
	}
	maxLevel := l.randomLevel()
	elem = &Element[K, V]{
		key:   key,
		value: value,
		lanes: make([]*Element[K, V], maxLevel),
		prev:  l.last,
	}
	lanes := l.header
	var next *Element[K, V]
	for level := l.maxLevel - 1; level >= 0; level-- {
		next = lanes[level]
		for next != nil && next.key < key {
			lanes, next = next.lanes, lanes[level]
		}
		if level < maxLevel {
			elem.lanes[level] = lanes[level]
			lanes[level] = elem
			if level == 0 {
				if elem.lanes[0] != nil {
					elem.prev = elem.lanes[0].prev
					elem.lanes[0].prev = elem
				}
			}
		}
	}
	if elem.Next() == nil {
		l.last = elem
	}
	if l.elements != nil {
		l.elements[key] = elem
	}
	l.length++
	return elem
}

// Remove a key-value pair from the list. If the key exists
// and is removed its element is returned. Note that this element
// is no longer in the list and thus has no Next() element.
// If the key was not found nil is returned instead.
// Complexity: O(log(n))
func (l *SkipList[K, V]) Remove(key K) (elem *Element[K, V]) {
	if l.rwMutex != nil {
		l.rwMutex.Lock()
		defer l.rwMutex.Unlock()
	}
	if l.elements != nil {
		if _, ok := l.elements[key]; !ok {
			return
		}
	}
	lanes := l.header
	var next *Element[K, V]
	for level := l.maxLevel - 1; level >= 0; level-- {
		next = lanes[level]
		for next != nil && next.key < key {
			lanes, next = next.lanes, lanes[level]
		}
		if next != nil && next.key == key {
			lanes[level] = next.lanes[level]
			if level == 0 {
				elem, next = next, next.Next()
				if next != nil {
					next.prev = elem.prev
				}
			}
		}
	}
	if elem != nil {
		l.length--

		// make sure to remove Next()
		elem.lanes[0] = nil

		if l.elements != nil {
			delete(l.elements, key)
		}
		if l.last.key == key {
			l.last = l.last.Prev()
		}
	}
	return
}

// Complexity: O(1)
func (l *SkipList[K, V]) RemoveFirst() (removed *Element[K, V]) {
	if l.rwMutex != nil {
		l.rwMutex.Lock()
		defer l.rwMutex.Unlock()
	}
	removed = l.First()
	if removed == nil {
		return
	}
	for level, elem := range l.header {
		if elem != nil && elem.key == removed.key {
			l.header[level] = removed.lanes[level]
		}
	}
	if l.elements != nil {
		delete(l.elements, removed.key)
	}
	l.length--
	if l.length == 0 {
		l.last = nil
	}
	if removed.lanes[0] != nil {
		removed.lanes[0].prev = nil
	}
	// remove Next()
	removed.lanes[0] = nil
	return
}

// Returns true if the key exists in the skip list, else false.
// Complexity: O(1) with hashmap, else O(log(n))
func (l *SkipList[K, V]) Contains(key K) bool {
	if l.rwMutex != nil {
		l.rwMutex.RLock()
		defer l.rwMutex.RUnlock()
	}
	return l.get(key) != nil
}

// Returns the element for a given key or nil
// if no element exists for the given key.
// Complexity: O(1) with hashmap, else O(log(n))
func (l *SkipList[K, V]) Get(key K) (elem *Element[K, V]) {
	if l.rwMutex != nil {
		l.rwMutex.RLock()
		defer l.rwMutex.RUnlock()
	}
	return l.get(key)
}

// Returns the element immediately preceeding
// the location of the given key.
// If no elements were found matching this condition
// then nil is returned.
// Complexity: O(log(n))
func (l *SkipList[K, V]) Before(key K) (elem *Element[K, V]) {
	elem = l.AtOrBefore(key)
	if elem != nil && elem.key == key {
		elem = elem.Prev()
	}
	return
}

// Given a key, returns the element of the first key that is larger
// than the given key.
// If no elements were found matching this condition nil is returned.
// Complexity: O(log(n))
func (l *SkipList[K, V]) After(key K) (elem *Element[K, V]) {
	elem = l.AtOrAfter(key)
	if elem != nil && elem.key == key {
		elem = elem.Next()
	}
	return
}

// Given a key, returns the element at that key or the closest preceeding key.
// If no element was found matching this condition nil is returned instead.
// An element at the given key (if it exists) is prioritized over the
// element of any key preceeding the given key.
// Complexity: O(log(n))
func (l *SkipList[K, V]) AtOrBefore(key K) (elem *Element[K, V]) {
	if l.rwMutex != nil {
		l.rwMutex.RLock()
		defer l.rwMutex.RUnlock()
	}
	elem = l.fastGet(key)
	if elem != nil {
		return elem
	}
	lanes := l.header
	for level := l.maxLevel - 1; level >= 0; level-- {
		if lanes[level] != nil && lanes[level].key <= key {
			elem = lanes[level]
			lanes = elem.lanes
			level++
		}
	}
	return
}

// Given a key, returns the element at that key or the closest succeeding key.
// If no element was found matching this condition nil is returned instead.
// An element at the given key (if it exists) nis prioritized over the value
// of any key succeeding the given key.
// Complexity: O(log(n))
func (l *SkipList[K, V]) AtOrAfter(key K) (elem *Element[K, V]) {
	if l.rwMutex != nil {
		l.rwMutex.RLock()
		defer l.rwMutex.RUnlock()
	}
	if l.header[0] != nil && l.header[0].key >= key {
		return l.header[0]
	}
	elem = l.fastGet(key)
	if elem != nil {
		return elem
	}
	lanes := l.header
	for level := l.maxLevel - 1; level >= 0; level-- {
		if lanes[level] != nil && lanes[level].key <= key {
			elem = lanes[level]
			lanes = elem.lanes
			level++
		}
	}
	if elem != nil && elem.key < key {
		elem = elem.lanes[0]
	}
	return
}

// Returns the element at a given key or nil
// if no element exists for the given key.
// Private function that does not lock any
// potential mutex.
// Complexity: O(1) with hashmap, O(log(n)) without
func (l *SkipList[K, V]) get(key K) (elem *Element[K, V]) {
	if l.elements != nil {
		return l.elements[key]
	}
	lanes := l.header
	for level := l.maxLevel - 1; level >= 0; level-- {
		if lanes[level] != nil && lanes[level].key <= key {
			elem = lanes[level]
			lanes = elem.lanes
			level++
		}
	}
	if elem != nil && elem.key != key {
		elem = nil
	}
	return
}

// Returns the element at a given key if hashmap is
// not disabled and the element exists, else returns nil.
// Private function that does not lock any
// potential mutex.
func (l *SkipList[K, V]) fastGet(key K) (elem *Element[K, V]) {
	if l.elements != nil {
		elem = l.elements[key]
	}
	return
}

// returns a level in the range [1, SkipList.maxLevel]
func (l *SkipList[K, V]) randomLevel() int {
	maxLevel := 1
	if l.prob == 0.5 && l.maxLevel <= 65 {
		// gotta go fast!
		result := (^uint64(0) >> (65 - uint64(l.maxLevel))) & l.rng.Uint64()
		for ; result&1 == 1; result >>= 1 {
			maxLevel++
		}
	} else {
		for ; maxLevel < l.maxLevel && l.rng.Float64() < l.prob; maxLevel++ {
		}
	}
	return maxLevel
}

type skipListOptions struct {
	maxLevel   int
	prob       float64
	seed       int64
	threadSafe bool
	hashmap    bool
}

func (o skipListOptions) validate() error {
	if o.maxLevel < 1 {
		return fmt.Errorf(
			"maximum level for skip list must be a positive integer, got %d",
			o.maxLevel,
		)
	}
	if o.prob < 0 || o.prob > 1 {
		return fmt.Errorf(
			"probability for skip list must be a floating point value in the range [0, 1], got %g",
			o.prob,
		)
	}
	return nil
}

type SkipListOption interface {
	apply(*skipListOptions)
}

var _ SkipListOption = (*withMaxLevel)(nil)

type withMaxLevel struct {
	maxLevel int
}

func (o *withMaxLevel) apply(opts *skipListOptions) {
	opts.maxLevel = o.maxLevel
}

// Set the max level of inserted elements.
// Without this option the max level is set
// to 32
func WithMaxLevel(maxLevel int) SkipListOption {
	return &withMaxLevel{maxLevel: maxLevel}
}

var _ SkipListOption = (*withSeed)(nil)

type withSeed struct {
	seed int64
}

func (o *withSeed) apply(opts *skipListOptions) {
	opts.seed = o.seed
}

// Seed the random number generator used
// for picking the levels for all new elements.
// Without this option the seed defaults to the
// current unix time in nanoseconds.
func WithSeed(seed int64) SkipListOption {
	return &withSeed{seed: seed}
}

var _ SkipListOption = (*withProbability)(nil)

type withProbability struct {
	prob float64
}

func (o *withProbability) apply(opts *skipListOptions) {
	opts.prob = o.prob
}

// Set the probability used for coinflips.
// Without this option the probability defaults to 0.5
func WithProbability(prob float64) SkipListOption {
	return &withProbability{prob: prob}
}

var _ SkipListOption = (*withThreadSafe)(nil)

type withThreadSafe struct{}

func (o *withThreadSafe) apply(opts *skipListOptions) {
	opts.threadSafe = true
}

// Use a RW mutex in the list.
// Without this option the
// threadsafety is turned off.
func WithThreadSafe() SkipListOption {
	return &withThreadSafe{}
}

var _ SkipListOption = (*withoutHashmap)(nil)

type withoutHashmap struct{}

// apply implements SkipListOption
func (*withoutHashmap) apply(opts *skipListOptions) {
	opts.hashmap = false
}

// Disable the use of hashmap, at times increasing
// lookup time. Reduces memory footprint.
func WithoutHashmap() SkipListOption {
	return &withThreadSafe{}
}
