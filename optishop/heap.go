package optishop

type MinHeap struct {
	nodes []*minHeapNode
}

func NewMinHeap() *MinHeap {
	return &MinHeap{nodes: []*minHeapNode{}}
}

func (m *MinHeap) Len() int {
	return len(m.nodes)
}

func (m *MinHeap) Push(data interface{}, priority float64) {
	m.nodes = append(m.nodes, &minHeapNode{data, priority})
	loc := heapLocationForIndex(len(m.nodes) - 1)
	// TODO: bubble up the heap as necessary.
}

func (m *MinHeap) Dequeue() (data interface{}, priority float64) {
	if m.Len() == 0 {
		return nil, nil
	}
	resNode := m.nodes[0]
	m.nodes[0] = m.nodes[len(m.nodes)-1]
	m.nodes = m.nodes[:len(m.nodes)-1]
	// TODO: push the new root node down as necessary.
}

type minHeapNode struct {
	Data     interface{}
	Priority float64
}

// A heapLocation identifies a location in a binary tree.
type heapLocation struct {
	depth int

	// The index of the heap node within the given depth
	// of the tree.
	localIdx int
}

func heapLocationForIndex(idx int) heapLocation {
	for depth := 0; true; depth++ {
		if idxInFlatTree(depth+1, 0) > idx {
			return heapLocation{
				depth:    depth,
				localIdx: idx - idxInFlatTree(depth, 0),
			}
		}
	}
	panic("unreachable")
}

func (h heapLocation) Index() int {
	return idxInFlatTree(h.depth, h.localIdx)
}

func idxInFlatTree(depth, idx int) int {
	if depth == 0 {
		return idx
	}
	return (1 << uint(depth-1)) + idx
}
