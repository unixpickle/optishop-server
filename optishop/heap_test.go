package optishop

import (
	"math/rand"
	"sort"
	"testing"
)

func TestHeapLocation(t *testing.T) {
	if (heapLocation{Depth: 3, LocalIdx: 3}).Index() != 10 {
		t.Error("unexpected index")
	}
	if heapLocationRoot().Index() != 0 {
		t.Error("unexpected index")
	}
	for i := 0; i < 1000; i++ {
		if heapLocationForIndex(i).Index() != i {
			t.Error("unexpected index")
		}
	}
}

func TestMinHeapOperations(t *testing.T) {
	heap := NewMinHeap()
	values := []float64{}
	for i := 0; i < 10000; i++ {
		n := rand.Intn(3)
		doInsert := n == 0
		doReplace := n == 1 && heap.Len() > 0
		if doInsert {
			value := float64(rand.Intn(1000))
			heap.Push(value*2, value)
			values = append(values, value*2)
		} else if doReplace {
			node := heap.nodes[rand.Intn(len(heap.nodes))]
			oldPriority := node.Priority
			value := float64(rand.Intn(1000))
			heap.Replace(node, value*2, value)
			for j, x := range values {
				if x == oldPriority*2 {
					values[j] = value * 2
					break
				}
			}
		} else {
			node := heap.Pop()
			if len(values) == 0 {
				if node != nil {
					t.Fatal("expected nil but got value")
				}
			} else {
				if node.Priority*2 != node.Data {
					t.Fatal("invalid value/priority pair")
				}
				sort.Float64s(values)
				if node.Data != values[0] {
					t.Fatal("unexpected minimimum value")
				}
				values = values[1:]
			}
		}
		testMinHeapInvariants(t, heap)
	}
}

func testMinHeapInvariants(t *testing.T, m *MinHeap) {
	for i := range m.nodes {
		loc := heapLocationForIndex(i)
		priority := m.nodes[loc.Index()].Priority
		child1, child2 := loc.Children()
		for _, child := range []heapLocation{child1, child2} {
			if child.Index() < len(m.nodes) {
				if m.nodes[child.Index()].Priority < priority {
					t.Fatal("child is lower priority than parent")
					return
				}
			}
		}
	}
}
