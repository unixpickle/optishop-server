package optishop

// A MinHeap implements a priority queue which selects the
// lowest priorities first.
type MinHeap struct {
	nodes []*MinHeapNode
}

// NewMinHeap creates an empty heap.
func NewMinHeap() *MinHeap {
	return &MinHeap{nodes: []*MinHeapNode{}}
}

// Len returns the number of elements in the heap.
func (m *MinHeap) Len() int {
	return len(m.nodes)
}

// Push adds an element to the heap, re-arranging the heap
// as necessary.
// Multiple values may be pushed with the same priority,
// in which case the order is undefined.
//
// The returned node can be passed to Replace() so long as
// the node remains in the heap.
func (m *MinHeap) Push(data interface{}, priority float64) *MinHeapNode {
	node := &MinHeapNode{len(m.nodes), data, priority}
	m.nodes = append(m.nodes, node)
	loc := heapLocationForIndex(node.index)
	m.moveUp(loc)
	return node
}

func (m *MinHeap) moveUp(loc heapLocation) {
	if loc.Depth == 0 {
		return
	}
	child := loc.Index()
	parent := loc.Parent().Index()
	if m.nodes[parent].Priority > m.nodes[child].Priority {
		m.swapNodes(parent, child)
		m.moveUp(loc.Parent())
	}
}

// Pop removes and returns (one of) the lowest priority
// elements in the queue.
//
// Returns nil if the heap is empty.
func (m *MinHeap) Pop() *MinHeapNode {
	if m.Len() == 0 {
		return nil
	}
	resNode := m.nodes[0]
	m.nodes[0] = m.nodes[len(m.nodes)-1]
	m.nodes[0].index = 0
	m.nodes = m.nodes[:len(m.nodes)-1]
	m.moveDown(heapLocationRoot())
	resNode.index = -1
	return resNode
}

func (m *MinHeap) moveDown(loc heapLocation) {
	child1, child2 := loc.Children()
	hasChild1 := child1.Index() < len(m.nodes)
	hasChild2 := child2.Index() < len(m.nodes)
	if !hasChild1 {
		// Reached the bottom of the tree.
	} else if hasChild1 && !hasChild2 {
		if m.nodes[child1.Index()].Priority < m.nodes[loc.Index()].Priority {
			m.swapNodes(child1.Index(), loc.Index())
			m.moveDown(child1)
		}
	} else {
		p := m.nodes[loc.Index()].Priority
		p1 := m.nodes[child1.Index()].Priority
		p2 := m.nodes[child2.Index()].Priority
		if p1 < p || p2 < p {
			other := child1
			if p2 < p1 {
				other = child2
			}
			m.swapNodes(other.Index(), loc.Index())
			m.moveDown(other)
		}
	}
}

// Replace updates the heap node to have a new value and
// priority, and then re-arranges the heap as necessary.
// This may only be called if n was returned by Push() and
// was not subsequently popped yet.
func (m *MinHeap) Replace(n *MinHeapNode, data interface{}, priority float64) {
	if n.index < 0 {
		panic("node is not in the heap")
	}
	oldPriority := n.Priority
	n.Data = data
	n.Priority = priority
	if priority < oldPriority {
		m.moveUp(heapLocationForIndex(n.index))
	} else if priority > oldPriority {
		m.moveDown(heapLocationForIndex(n.index))
	}
}

func (m *MinHeap) swapNodes(idx1, idx2 int) {
	m.nodes[idx1], m.nodes[idx2] = m.nodes[idx2], m.nodes[idx1]
	m.nodes[idx1].index = idx1
	m.nodes[idx2].index = idx2
}

type MinHeapNode struct {
	index int

	Data     interface{}
	Priority float64
}

// A heapLocation identifies a location in a binary tree.
type heapLocation struct {
	Depth int

	// The index of the heap node within the given depth
	// of the tree.
	LocalIdx int
}

func heapLocationRoot() heapLocation {
	return heapLocation{}
}

func heapLocationForIndex(idx int) heapLocation {
	for depth := 0; true; depth++ {
		if idxInFlatTree(depth+1, 0) > idx {
			return heapLocation{
				Depth:    depth,
				LocalIdx: idx - idxInFlatTree(depth, 0),
			}
		}
	}
	panic("unreachable")
}

func (h heapLocation) Index() int {
	return idxInFlatTree(h.Depth, h.LocalIdx)
}

func (h heapLocation) Parent() heapLocation {
	return heapLocation{Depth: h.Depth - 1, LocalIdx: h.LocalIdx / 2}
}

func (h heapLocation) Children() (heapLocation, heapLocation) {
	return heapLocation{Depth: h.Depth + 1, LocalIdx: h.LocalIdx * 2},
		heapLocation{Depth: h.Depth + 1, LocalIdx: h.LocalIdx*2 + 1}
}

func idxInFlatTree(depth, idx int) int {
	if depth == 0 {
		return 0
	}
	return (1 << uint(depth)) + idx - 1
}
