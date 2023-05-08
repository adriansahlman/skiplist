// Package skiplist implements the skip list datastructure.
package skiplist

import (
	"math/rand"
	"time"

	"golang.org/x/exp/constraints"
)

const (
	MaxLevel = 32
)

type Node[K, V any] struct {
	key   K
	value V

	// The skip-lanes. lanes[0] refers to
	// the node directly succeeding this
	// node in the list.
	lanes []*Node[K, V]
	// The node directly preceeding this node
	// in the list.
	prev *Node[K, V]
}

func (n *Node[K, V]) Key() K {
	return n.key
}

// Returns the value of the node.
func (n *Node[K, V]) Value() V {
	return n.value
}

// Returns the next node or nil if no next node exists.
func (n *Node[K, V]) Next() *Node[K, V] {
	return n.lanes[0]
}

// Returns the previous node or nil if no previous node exists.
func (n *Node[K, V]) Prev() *Node[K, V] {
	return n.prev
}

// Implements a skip-list (doubly linked) with
// a maximum level of 32.
type SkipList[K constraints.Ordered, V any] struct {
	// fast lookup of exact key
	nodes map[K]*Node[K, V]
	// skip-lanes
	header []*Node[K, V]
	// the last node in the list
	last *Node[K, V]
	// number of nodes
	length int
	// random number generator used for
	// selecting the level of new nodes.
	rng *rand.Rand
}

// Create a new doubly-linked skiplist.
func New[K constraints.Ordered, V any](
	opts ...Option,
) *SkipList[K, V] {
	o := skipListOptions{
		seed: time.Now().UnixNano(),
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	var nodes map[K]*Node[K, V]
	if o.hashmap {
		nodes = make(map[K]*Node[K, V])
	}
	return &SkipList[K, V]{
		nodes:  nodes,
		header: make([]*Node[K, V], MaxLevel),
		rng:    rand.New(rand.NewSource(o.seed)),
	}
}

// Returns the number of nodes in the list.
func (l *SkipList[K, V]) Length() int {
	return l.length
}

// Set the value for a key. Returns the node containing
// the key and new value.
// Average complexity: O(log(n))
func (l *SkipList[K, V]) Set(
	key K,
	value V,
) *Node[K, V] {
	var node *Node[K, V]

	// if using a hashmap and a node with the
	// key already exists we can simply replace
	// its value and return without having to
	// traverse the skip list.
	if l.nodes != nil {
		if node = l.nodes[key]; node != nil {
			node.value = value
			return node
		}
	}

	// in range [1, 32]
	nodeLevel := 1 + sampleGeometricDistribution(MaxLevel-1, l.rng)

	node = &Node[K, V]{
		key:   key,
		value: value,
		lanes: make([]*Node[K, V], nodeLevel),
	}

	// track if we are replacing an existing value
	// in which case we dont need to increment the
	// list length.
	var replaced bool

	lanes := l.header
	for level := MaxLevel - 1; level >= 0; level-- {
		for ; lanes[level] != nil && lanes[level].key < key; lanes = lanes[level].lanes {
		}
		if lanes[level] != nil && lanes[level].key == key {
			replaced = true
			// route around existing node, removing
			// any references to it for the current lane.
			node.prev = lanes[level].prev
			lanes[level] = lanes[level].lanes[level]
		}
		if level < nodeLevel {
			node.lanes[level] = lanes[level]
			lanes[level] = node
			if level == 0 && node.lanes[0] != nil {
				if !replaced {
					// prev for the new node has
					// not been set yet.
					node.prev = node.lanes[0].prev
				}
				// prev for the next node should
				// point back to the new node.
				node.lanes[0].prev = node
			}
		}
	}
	if !replaced {
		l.length++
	}
	if l.nodes != nil {
		l.nodes[key] = node
	}
	if l.last == nil || l.last.key < key {
		node.prev = l.last
		l.last = node
	}
	return node
}

// Get the node for a key.
// Returns nil if no node exists for the given key.
// Average complexity: O(log(n))
// When hashmap is enabled, the complexity is
// O(1) instead.
func (l *SkipList[K, V]) Get(key K) *Node[K, V] {
	if l.nodes != nil {
		return l.nodes[key]
	}
	if node := l.Search(key); node != nil && node.key == key {
		return node
	}
	return nil
}

// Get the first node in the list.
// Returns nil if list is empty.
// Complexity: O(1)
func (l *SkipList[K, V]) First() *Node[K, V] {
	return l.header[0]
}

// Get the last node in the list.
// Returns nil if list is empty.
// Complexity: O(1)
func (l *SkipList[K, V]) Last() *Node[K, V] {
	return l.last
}

// Remove a node from the list with a given key
// and return it. Returns nil if no node was found
// for the given key.
// Average complexity: O(log(n))
func (l *SkipList[K, V]) Remove(
	key K,
) *Node[K, V] {
	// when using a hashmap we can quickly
	// check if the given key exists. If we
	// know it does not exist in the list we
	// can return early as there is nothing
	// to remove.
	if l.nodes != nil {
		if _, ok := l.nodes[key]; !ok {
			return nil
		}
	}
	lanes := l.header
	var node *Node[K, V]
	for level := MaxLevel - 1; level >= 0; level-- {
		for ; lanes[level] != nil && lanes[level].key < key; lanes = lanes[level].lanes {
		}
		if lanes[level] != nil && lanes[level].key == key {
			// grab the node being removed
			node = lanes[level]
			// route forward lane to the node succeeding
			// the node being removed for the current level.
			lanes[level] = lanes[level].lanes[level]
		}
	}

	// was a node found for the key and removed?
	if node != nil {
		l.length--
		if l.nodes != nil {
			delete(l.nodes, key)
		}
		if node.lanes[0] == nil {
			l.last = node.prev
		} else {
			// route backward lane to the node preceeding
			// the node being removed.
			node.lanes[0].prev = node.prev
		}
	}
	return node
}

// Remove the first node in the list and return it.
// Returns nil if the list is empty.
// Complexity: O(1)
func (l *SkipList[K, V]) RemoveFirst() *Node[K, V] {
	node := l.header[0]
	if node == nil {
		return nil
	}
	// route the forward lanes around the node
	// being removed.
	for level := range l.header {
		if l.header[level] == node {
			l.header[level] = node.lanes[level]
		}
	}
	if l.nodes != nil {
		delete(l.nodes, node.key)
	}
	l.length--
	if l.length == 0 {
		l.last = nil
	} else if node.lanes[0] != nil {
		// we know that no previous node exists
		// for the new first node in the list as
		// we just removed its preceeding node.
		node.lanes[0].prev = nil
	}
	return node
}

// Find the node at a key or the node with the smallest key
// that is larger or equal to the given key.
// Returns nil if no node exists with a key that is larger
// or equal to the given key.
// Average complexity: O(log(n))
func (l *SkipList[K, V]) Search(
	key K,
) *Node[K, V] {
	var node *Node[K, V]
	if l.nodes != nil {
		node = l.nodes[key]
		if node != nil {
			return node
		}
	}
	lanes := l.header
	for level := MaxLevel - 1; level >= 0; level-- {
		for ; lanes[level] != nil && lanes[level].key < key; lanes = lanes[level].lanes {
		}
	}
	return lanes[0]
}

type skipListOptions struct {
	seed    int64
	hashmap bool
}

type Option interface {
	apply(*skipListOptions)
}

var _ Option = (*withSeed)(nil)

type withSeed struct {
	seed int64
}

func (o *withSeed) apply(opts *skipListOptions) {
	opts.seed = o.seed
}

// Seed the random number generator used
// for picking the levels for all new nodes.
// Without this option the seed defaults to the
// current unix time in nanoseconds.
func WithSeed(seed int64) Option {
	return &withSeed{seed: seed}
}

var _ Option = (*withHashmap)(nil)

type withHashmap struct{}

// apply implements SkipListOption
func (*withHashmap) apply(opts *skipListOptions) {
	opts.hashmap = true
}

// Enable the use of hashmap, reducing
// lookup time when a given key already
// exists in the list. Increases memory
// footprint.
// Note that the Set(key, value) function
// of the skiplist behaves slightly different
// when a hashmap is in use. If a node with
// the given key already exists, the nodes
// value is simply replaced. Without a hashmap,
// a new node is created with a new random level
// which completely replaces the any previous
// node with the same key.
func WithHashmap() Option {
	return &withHashmap{}
}

// Sample from a geometric distribution with
// a probability of 0.5. The maxmimum returned
// value is min(32, limit).
func sampleGeometricDistribution(
	limit int,
	rng *rand.Rand,
) int {
	var n int
	result := (^uint32(0) >> (32 - uint32(limit))) & rng.Uint32()
	for ; result&1 == 1; result >>= 1 {
		n++
	}
	return n
}
