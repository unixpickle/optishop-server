package optishop

import (
	"math"

	"github.com/unixpickle/essentials"
)

// A FloorPoint is a Point within a specific floor.
// The floor is specified by its index within a Layout.
type FloorPoint struct {
	Point

	// Floor stores the floor index.
	Floor int
}

// A FloorConnector combines Connectors from each floor
// of a building to connect arbitrary locations in the
// building.
type FloorConnector struct {
	Connectors []Connector
	Layout     *Layout
}

// NewFloorConnector creates a new FloorConnector using
// Rasters for the Connector implementation.
func NewFloorConnector(layout *Layout) *FloorConnector {
	conn := &FloorConnector{
		Layout:     layout,
		Connectors: make([]Connector, len(layout.Floors)),
	}
	for i, floor := range layout.Floors {
		conn.Connectors[i] = NewRaster(floor)
	}
	return conn
}

// NewFloorConnectorCached is like NewFloorConnector, but
// the internal connectors cache results and use batch
// computations so that computing all pairwise distances
// between N points is still O(N).
func NewFloorConnectorCached(layout *Layout) *FloorConnector {
	conn := &FloorConnector{
		Layout:     layout,
		Connectors: make([]Connector, len(layout.Floors)),
	}
	for i, floor := range layout.Floors {
		conn.Connectors[i] = NewCacheConnector(NewRaster(floor))
	}
	return conn
}

// Connect finds a short FloorPath between points a and b.
//
// Returns nil if no path could be found.
func (f *FloorConnector) Connect(a, b FloorPoint) FloorPath {
	portalDistance := f.portalDistance()

	queue := NewMinHeap()
	visited := map[FloorPoint]*MinHeapNode{}

	addNode := func(n *floorSearchNode, dist float64) {
		fp := n.FloorPoint(f.Layout)
		if oldNode, ok := visited[fp]; ok {
			if oldNode.Priority > dist {
				queue.Replace(oldNode, n, dist)
			}
		} else {
			visited[fp] = queue.Push(n, dist)
		}
	}

	expandNode := func(prev *floorSearchNode, dist float64, fp FloorPoint) {
		conn := f.Connectors[fp.Floor]

		if fp.Floor == b.Floor {
			path := conn.Connect(fp.Point, b.Point)
			step := &FloorPathStep{
				Floor: fp.Floor,
				Path:  path,
			}
			node := &floorSearchNode{Final: true, Step: step, Parent: prev}
			addNode(node, dist+path.Length())
			return
		}

		for _, portal := range f.Layout.Floors[fp.Floor].Portals {
			path := conn.Connect(fp.Point, portal.Location)
			for _, dest := range portal.Destinations {
				step := &FloorPathStep{
					Floor:        fp.Floor,
					Path:         path,
					SourcePortal: portal.ID,
					DestPortal:   dest,
				}
				node := &floorSearchNode{Step: step, Parent: prev}
				addNode(node, dist+path.Length()+portalDistance)
			}
		}
	}

	expandNode(nil, 0, a)
	for queue.Len() > 0 {
		rawNode := queue.Pop()
		node := rawNode.Data.(*floorSearchNode)
		if node.Final {
			path := FloorPath{}
			for node != nil {
				path = append(path, node.Step)
				node = node.Parent
			}
			essentials.Reverse(path)
			return path
		}
		expandNode(node, rawNode.Priority, node.FloorPoint(f.Layout))
	}

	return nil
}

// DistanceFunc creates a function that computes distances
// between any two locations in a list of locations.
//
// This distance function can be used with a TSP solver.
func (f *FloorConnector) DistanceFunc(points []FloorPoint) func(idx1, idx2 int) float64 {
	portalDist := f.portalDistance()
	distances := make([][]float64, len(points))
	for i, p := range points {
		distances[i] = make([]float64, len(points))
		for j, p1 := range points {
			if i == j {
				continue
			}
			path := f.Connect(p, p1)
			distances[i][j] = float64(len(path)) * portalDist
			for _, part := range path {
				distances[i][j] += part.Path.Length()
			}
		}
	}
	return func(i, j int) float64 {
		return distances[i][j]
	}
}

// portalDistance gets a relatively long distance that can
// be used to represent going through a portal.
// This distance is intended to be long enough that
// portals will always be considered more expensive than
// walking within a given floor.
func (f *FloorConnector) portalDistance() float64 {
	largeDistance := 0.0
	for _, floor := range f.Layout.Floors {
		_, _, w, h := floor.Bounds.Bounds()
		largeDistance = math.Max(largeDistance, math.Max(w, h))
	}
	return largeDistance * 100
}

// A FloorPathStep is a single step in a path between two
// arbitrary places in a building.
//
// Every step except for the last step takes the user
// through a portal.
type FloorPathStep struct {
	// Floor specifies the floor on which the path takes
	// place.
	Floor int

	// Path is the path within the given floor.
	Path Path

	// These fields specify how a portal is used at the
	// end of this step.
	// Unused for the final step of a FloorPath.
	SourcePortal int
	DestPortal   int
}

// A FloorPath is a set of steps that takes the user from
// one point in a store to another.
type FloorPath []*FloorPathStep

type floorSearchNode struct {
	Final  bool
	Step   *FloorPathStep
	Parent *floorSearchNode
}

func (f *floorSearchNode) FloorPoint(l *Layout) FloorPoint {
	if f.Final {
		return FloorPoint{Point: f.Step.Path[len(f.Step.Path)-1], Floor: f.Step.Floor}
	} else {
		destPortal := l.Portal(f.Step.DestPortal)
		return FloorPoint{
			Point: destPortal.Location,
			Floor: l.PortalFloor(destPortal),
		}
	}
}
