package skiplist

import (
	"math/rand"
)

const MaxLevel = 32

// Create a new skiplist.
func New[T any](
	less func(a, b T) bool,
	opts ...Option,
) *SkipList[T] {
	o := options{}
	for _, opt := range opts {
		opt.apply(&o)
	}
	if o.rng == nil {
		o.rng = rand.New(rand.NewSource(0)).Uint32
	}
	return &SkipList[T]{
		lanes:   make([]*Node[T], MaxLevel),
		less:    less,
		replace: o.replace,
		rng:     o.rng,
	}
}

type options struct {
	rng     func() uint32
	replace bool
}

type SkipList[T any] struct {
	less    func(a, b T) bool
	lanes   []*Node[T]
	last    *Node[T]
	length  int
	replace bool
	rng     func() uint32
}

// Returns the number of nodes in the skiplist.
func (l *SkipList[T]) Length() int {
	return l.length
}

// Clear the contents of the skiplist, setting
// its length to 0.
func (l *SkipList[T]) Clear() {
	for i := range l.lanes {
		l.lanes[i] = nil
	}
	l.last = nil
	l.length = 0
}

// Get the first node in the skiplist.
// Returns nil if the skiplist is empty.
// Complexity: O(1)
func (l *SkipList[T]) First() *Node[T] {
	return l.lanes[0]
}

// Get the last node in the skiplist.
// Returns nil if the skiplist is empty.
// Complexity: O(1)
func (l *SkipList[T]) Last() *Node[T] {
	return l.last
}

// Insert a value into the skiplist and return its node.
// Average complexity: O(log(n))
func (l *SkipList[T]) Add(value T) (node *Node[T]) {
	level := 1
	// add geometric distribution sample in range [0, 31]
	for i := (^uint32(0) >> 1) & l.rng(); i&1 == 1; i >>= 1 {
		level++
	}
	node = &Node[T]{
		value: value,
		lanes: make([]*Node[T], level),
	}

	var nodeReplaced bool

	lanes := l.lanes
	if l.replace {
		for levelIdx := MaxLevel - 1; levelIdx >= 0; levelIdx-- {
			for ; lanes[levelIdx] != nil && l.less(lanes[levelIdx].value, value); lanes = lanes[levelIdx].lanes {
			}
			if lanes[levelIdx] != nil && !l.less(value, lanes[levelIdx].value) {
				nodeReplaced = true
				// route around existing node, removing
				// any references to it for the current lane.
				node.prev = lanes[levelIdx].prev
				lanes[levelIdx] = lanes[levelIdx].lanes[levelIdx]
			}
			if levelIdx < level {
				node.lanes[levelIdx] = lanes[levelIdx]
				lanes[levelIdx] = node
				if levelIdx == 0 && node.lanes[0] != nil {
					if !nodeReplaced {
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
	} else {
		for levelIdx := MaxLevel - 1; levelIdx >= 0; levelIdx-- {
			for ; lanes[levelIdx] != nil && l.less(lanes[levelIdx].value, value); lanes = lanes[levelIdx].lanes {
			}
			if levelIdx >= level {
				continue
			}
			node.lanes[levelIdx] = lanes[levelIdx]
			lanes[levelIdx] = node
			if levelIdx == 0 && node.lanes[0] != nil {
				// prev for the new node has
				// not been set yet.
				node.prev = node.lanes[0].prev
				// prev for the next node should
				// point back to the new node.
				node.lanes[0].prev = node
			}
		}
	}
	if !nodeReplaced {
		l.length++
	}
	if l.last == nil || l.less(l.last.value, value) {
		node.prev = l.last
		l.last = node
	}
	return
}

// Find and return the first node with a value that is
// greater or equal to the given value.
// Returns nil if no such node exists.
// Average complexity: O(log(n))
func (l *SkipList[T]) Search(
	value T,
) (node *Node[T]) {
	lanes := l.lanes
	for levelIdx := MaxLevel - 1; levelIdx >= 0; levelIdx-- {
		for ; lanes[levelIdx] != nil && l.less(lanes[levelIdx].value, value); lanes = lanes[levelIdx].lanes {
		}
	}
	return lanes[0]
}

// Remove the first node encountered for a given value
// and return it.
// Returns nil if no node with the value was found.
// Average complexity: O(log(n))
func (l *SkipList[T]) Remove(
	value T,
) (node *Node[T]) {
	lanes := l.lanes
	for levelIdx := MaxLevel - 1; levelIdx >= 0; levelIdx-- {
		for ; lanes[levelIdx] != nil && l.less(lanes[levelIdx].value, value); lanes = lanes[levelIdx].lanes {
		}
		if lanes[levelIdx] != nil && !l.less(value, lanes[levelIdx].value) {
			// grab the node being removed
			node = lanes[levelIdx]
			// route forward lane to the node succeeding
			// the node being removed for the current level.
			lanes[levelIdx] = lanes[levelIdx].lanes[levelIdx]
		}
	}

	if node == nil {
		// node with given value was not found, return nothing
		return
	}
	l.length--
	if node.lanes[0] == nil {
		l.last = node.prev
	} else {
		// route backward lane to the node preceeding
		// the node being removed.
		node.lanes[0].prev = node.prev
	}
	return
}

// Remove the first node in the sorted collection and
// return it.
// Returns nil if the collection is empty.
// Complexity: O(1)
func (l *SkipList[T]) RemoveFirst() (node *Node[T]) {
	if node = l.lanes[0]; node == nil {
		return
	}
	// route the forward lanes around the node
	// being removed.
	for levelIdx := range l.lanes {
		if l.lanes[levelIdx] == node {
			l.lanes[levelIdx] = node.lanes[levelIdx]
		}
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
	return
}

type Node[T any] struct {
	value T
	// The next node and any optional skiplanes.
	lanes []*Node[T]
	// The node directly preceeding this node
	// in the list.
	prev *Node[T]
}

// Get the value of the node.
func (n *Node[T]) Value() T {
	return n.value
}

// Get the next node.
func (n *Node[T]) Next() *Node[T] {
	return n.lanes[0]
}

// Get the previous node.
func (n *Node[T]) Prev() *Node[T] {
	return n.prev
}

// Get the node level.
// The level is in the range [1, 32].
func (n *Node[T]) Level() int {
	return len(n.lanes)
}

// Remove any occurence of this node in the given skiplist.
// Returns itself if the node was found, else nil.
// Average complexity: O(log(n))
// If this is the first node in the skiplist its removal
// operation has a complexity of O(1).
func (n *Node[T]) RemoveFrom(
	l *SkipList[T],
) (node *Node[T]) {
	if n == nil {
		return
	}
	lanes := l.lanes
	if lanes[0] == n {
		return l.RemoveFirst()
	}
	for levelIdx := MaxLevel - 1; levelIdx >= 0; levelIdx-- {
		for ; lanes[levelIdx] != nil && l.less(lanes[levelIdx].value, n.value); lanes = lanes[levelIdx].lanes {
		}
		if !l.replace {
			// There may be more nodes that match the value
			// of the node being removed. The nodes are traversed
			// while node values are equal to the value of the node
			// being removed. Stops on node match.
			//
			// Revert to the current node if there
			// are no node matches.
			currentLanes := lanes
			for ; lanes[levelIdx] != nil && !l.less(n.value, lanes[levelIdx].value) && lanes[levelIdx] != n; lanes = lanes[levelIdx].lanes {
			}
			if lanes[levelIdx] != n {
				lanes = currentLanes
			}
		}
		if lanes[levelIdx] == n {
			// grab the node being removed
			node = lanes[levelIdx]
			// route forward lane to the node succeeding
			// the node being removed for the current level.
			lanes[levelIdx] = lanes[levelIdx].lanes[levelIdx]
		}
	}

	if node == nil {
		// node was not found, return nothing
		return
	}
	l.length--
	if node.lanes[0] == nil {
		l.last = node.prev
	} else {
		// route backward lane to the node preceeding
		// the node being removed.
		node.lanes[0].prev = node.prev
	}
	return
}

type Option interface {
	apply(*options)
}

var _ Option = (*withRng)(nil)

type withRng struct {
	rng func() uint32
}

func (o *withRng) apply(opts *options) {
	opts.rng = o.rng
}

// Use a custom random number generator.
func WithRng(rng func() uint32) Option {
	return &withRng{rng: rng}
}

var _ Option = (*withReplace)(nil)

type withReplace struct{}

func (o *withReplace) apply(opts *options) {
	opts.replace = true
}

// When adding a value (node) to the skiplist, remove
// any other nodes that hold the same value.
func WithReplace() Option {
	return &withReplace{}
}
