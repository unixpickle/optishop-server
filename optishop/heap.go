package optishop

// A MinHeap implements a priority queue which selects the
// lowest priorities first.
type MinHeap struct {
	nodes []*minHeapNode
}

// NewMinHeap creates an empty heap.
func NewMinHeap() *MinHeap {
	return &MinHeap{nodes: []*minHeapNode{}}
}

// Len returns the number of elements in the heap.
func (m *MinHeap) Len() int {
	return len(m.nodes)
}

// Push adds an element to the heap, re-arranging the heap
// as necessary.
// Multiple values may be pushed with the same priority,
// in which case the order is undefined.
func (m *MinHeap) Push(data interface{}, priority float64) {
	m.nodes = append(m.nodes, &minHeapNode{data, priority})
	loc := heapLocationForIndex(len(m.nodes) - 1)
	m.moveUp(loc)
}

func (m *MinHeap) moveUp(loc heapLocation) {
	if loc.Depth == 0 {
		return
	}
	child := loc.Index()
	parent := loc.Parent().Index()
	if m.nodes[parent].Priority > m.nodes[child].Priority {
		m.nodes[parent], m.nodes[child] = m.nodes[child], m.nodes[parent]
		m.moveUp(loc.Parent())
	}
}

// Pop removes and returns (one of) the lowest priority
// elements in the queue.
func (m *MinHeap) Pop() (data interface{}, priority float64) {
	if m.Len() == 0 {
		return nil, 0
	}
	resNode := m.nodes[0]
	m.nodes[0] = m.nodes[len(m.nodes)-1]
	m.nodes = m.nodes[:len(m.nodes)-1]
	m.moveDown(heapLocationRoot())
	return resNode.Data, resNode.Priority
}

func (m *MinHeap) moveDown(loc heapLocation) {
	child1, child2 := loc.Children()
	hasChild1 := child1.Index() < len(m.nodes)
	hasChild2 := child2.Index() < len(m.nodes)
	if !hasChild1 {
		// Reached the bottom of the tree.
	} else if hasChild1 && !hasChild2 {
		if m.nodes[child1.Index()].Priority < m.nodes[loc.Index()].Priority {
			m.nodes[child1.Index()], m.nodes[loc.Index()] = m.nodes[loc.Index()],
				m.nodes[child1.Index()]
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
			m.nodes[other.Index()], m.nodes[loc.Index()] = m.nodes[loc.Index()],
				m.nodes[other.Index()]
			m.moveDown(other)
		}
	}
}

type minHeapNode struct {
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
