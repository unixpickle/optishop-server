package optishop

import (
	"math"

	"github.com/unixpickle/essentials"
)

const (
	maxNearbyDelta = 4
	rasterSize     = 600
)

// A Connector finds short paths from one point to another
// on a Floor, avoiding obstacles as needed.
type Connector interface {
	// Obstructed checks if a point is obstructed.
	// A point is obstructed if it is either inside of an
	// obstacle, or outside of the floor's bounds.
	Obstructed(p Point) bool

	// Unobstruct gets a point close to p which is not
	// obstructed.
	Unobstruct(p Point) Point

	// NonPreferred checks if a point is in an area that
	// should only be entered if it contains a source or a
	// destination.
	NonPreferred(p Point) bool

	// Connect finds a path connecting points a and b.
	//
	// If the start or end point is obstructed, a nearby
	// unobstructed point is used.
	//
	// If no path can be found, nil is returned.
	Connect(a, b Point) Path
}

// NewConnector creates a concrete implementation of
// Connector for the given floor plan.
func NewConnector(f *Floor) Connector {
	// The raster size must preserve the same aspect ratio
	// so that diagonal lines are really the correct
	// length.
	_, _, w, h := f.Bounds.Bounds()
	if w > h {
		h *= rasterSize / w
		w = rasterSize
	} else {
		w *= rasterSize / h
		h = rasterSize
	}
	return newRasterConnector(f, int(math.Ceil(w)), int(math.Ceil(h)))
}

type rasterPoint struct {
	X int
	Y int
}

type rasterConnector struct {
	boundsX      float64
	boundsY      float64
	boundsWidth  float64
	boundsHeight float64

	width  int
	height int

	obstructed   []bool
	nonPreferred []bool
}

func newRasterConnector(floor *Floor, width, height int) *rasterConnector {
	x, y, w, h := floor.Bounds.Bounds()
	res := &rasterConnector{
		boundsX:      x,
		boundsY:      y,
		boundsWidth:  w,
		boundsHeight: h,

		width:  width,
		height: height,

		obstructed:   make([]bool, width*height),
		nonPreferred: make([]bool, width*height),
	}

	res.checkBoundaries(floor.Bounds)
	res.addToRaster(res.obstructed, floor.Obstacles)
	res.addToRaster(res.nonPreferred, floor.NonPreferred)

	return res
}

func (r *rasterConnector) Obstructed(p Point) bool {
	rp := r.pointToRaster(p)
	if rp.X < 0 || rp.Y < 0 || rp.X >= r.width || rp.Y >= r.height {
		return true
	}
	return r.obstructed[r.pointToIndex(rp)]
}

func (r *rasterConnector) NonPreferred(p Point) bool {
	rp := r.pointToRaster(p)
	if rp.X < 0 || rp.Y < 0 || rp.X >= r.width || rp.Y >= r.height {
		return false
	}
	return r.nonPreferred[r.pointToIndex(rp)]
}

func (r *rasterConnector) Unobstruct(p Point) Point {
	start := r.pointToRaster(p)
	start.X = clampDim(start.X, r.width)
	start.Y = clampDim(start.Y, r.height)
	// Basic BFS to find the nearest point in L1 distance.
	queue := []rasterPoint{start}
	visited := map[rasterPoint]bool{start: true}
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		if !r.obstructed[r.pointToIndex(next)] {
			return r.rasterToPoint(next)
		}
		for _, newPoint := range r.neighbors(next) {
			if !visited[newPoint] {
				visited[newPoint] = true
				queue = append(queue, newPoint)
			}
		}
	}
	// No unobstructed points, which should never happen.
	// Don't panic() here, because if this ever _does_
	// happen, connecting two points can fail gracefully
	// without a panic().
	return p
}

func (r *rasterConnector) Connect(a, b Point) Path {
	start := r.pointToRaster(r.Unobstruct(a))
	end := r.pointToRaster(r.Unobstruct(b))
	dists := newRasterDistances()

	maxPreferredChanges := 0
	if r.nonPreferred[r.pointToIndex(start)] {
		maxPreferredChanges++
	}
	if r.nonPreferred[r.pointToIndex(end)] {
		maxPreferredChanges++
	}

	queue := NewMinHeap()
	firstNode := queue.Push(&connectorSearchNode{
		Point:        start,
		NonPreferred: r.nonPreferred[r.pointToIndex(start)],
	}, 0)
	visited := make([]*MinHeapNode, r.width*r.height)
	visited[r.pointToIndex(start)] = firstNode

	for queue.Len() > 0 {
		queueNode := queue.Pop()
		distance := queueNode.Priority
		node := queueNode.Data.(*connectorSearchNode)

		if node.Point == end {
			points := Path{b}
			for node != nil {
				points = append(points, r.rasterToPoint(node.Point))
				node = node.Parent
			}
			points = append(points, a)
			essentials.Reverse(points)
			return points
		}

		r.nearbyPoints(node.Point, func(newPoint rasterPoint) {
			newDist := dists.Distance(node.Point, newPoint) + distance
			newIdx := r.pointToIndex(newPoint)
			if searchNode := visited[newIdx]; searchNode == nil || searchNode.Priority > newDist {
				newNode := node.AddStep(newPoint, r.nonPreferred[newIdx])
				if newNode.NumPreferredChanges <= maxPreferredChanges {
					if searchNode == nil {
						visited[newIdx] = queue.Push(newNode, newDist)
					} else {
						queue.Replace(searchNode, newNode, newDist)
					}
				}
			}
		})
	}

	return nil
}

func (r *rasterConnector) checkBoundaries(bounds Polygon) {
	idx := 0
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			p := r.rasterToPoint(rasterPoint{X: x, Y: y})
			if !bounds.Contains(p) {
				r.obstructed[idx] = true
			}
			idx++
		}
	}
}

func (r *rasterConnector) addToRaster(raster []bool, polygons []Polygon) {
	for _, poly := range polygons {
		x, y, width, height := poly.Bounds()
		minX := int(float64(r.width) * (x - r.boundsX) / r.boundsWidth)
		minY := int(float64(r.height) * (y - r.boundsY) / r.boundsHeight)
		maxX := int(math.Ceil(float64(r.width) * (x + width - r.boundsX) / r.boundsWidth))
		maxY := int(math.Ceil(float64(r.height) * (y + height - r.boundsY) / r.boundsHeight))
		minX = clampDim(minX, r.width)
		minY = clampDim(minY, r.height)
		maxX = clampDim(maxX, r.width)
		maxY = clampDim(maxY, r.height)
		for i := minY; i <= maxY; i++ {
			for j := minX; j <= maxX; j++ {
				rp := rasterPoint{X: j, Y: i}
				p := r.rasterToPoint(rp)
				if poly.Contains(p) {
					raster[r.pointToIndex(rp)] = true
				}
			}
		}
	}
}

func (r *rasterConnector) pointToRaster(p Point) rasterPoint {
	return rasterPoint{
		X: int(math.Round(float64(r.width) * (p.X - r.boundsX) / r.boundsWidth)),
		Y: int(math.Round(float64(r.height) * (p.Y - r.boundsY) / r.boundsHeight)),
	}
}

func (r *rasterConnector) rasterToPoint(p rasterPoint) Point {
	return Point{
		X: (float64(p.X)/float64(r.width))*r.boundsWidth + r.boundsX,
		Y: (float64(p.Y)/float64(r.height))*r.boundsHeight + r.boundsY,
	}
}

func (r *rasterConnector) pointToIndex(p rasterPoint) int {
	return p.X + p.Y*r.width
}

func (r *rasterConnector) neighbors(p rasterPoint) []rasterPoint {
	res := make([]rasterPoint, 0, 4)
	if p.X > 0 {
		res = append(res, rasterPoint{X: p.X - 1, Y: p.Y})
	}
	if p.Y > 0 {
		res = append(res, rasterPoint{X: p.X, Y: p.Y - 1})
	}
	if p.X+1 < r.width {
		res = append(res, rasterPoint{X: p.X + 1, Y: p.Y})
	}
	if p.Y+1 < r.height {
		res = append(res, rasterPoint{X: p.X, Y: p.Y + 1})
	}
	return res
}

// nearbyPoints finds a small square neighborhood of
// points around p that definitely are not blocked by
// obstacles and can be reached directly, and calls f for
// each such point.
func (r *rasterConnector) nearbyPoints(p rasterPoint, f func(rasterPoint)) {
	hitObstacle := false
	for delta := 1; delta <= maxNearbyDelta && !hitObstacle; delta++ {
		for i := -delta; i <= delta; i++ {
			for _, j := range []int{-delta, delta} {
				p1 := rasterPoint{X: p.X + i, Y: p.Y + j}
				p2 := rasterPoint{X: p.X + j, Y: p.Y + i}
				for _, rp := range []rasterPoint{p1, p2} {
					if rp.X < 0 || rp.Y < 0 || rp.X+1 >= r.width || rp.Y+1 >= r.height {
						continue
					}
					if r.obstructed[r.pointToIndex(rp)] {
						hitObstacle = true
					} else {
						f(rp)
					}
				}

			}
		}
	}
}

func clampDim(x, dim int) int {
	if x < 0 {
		return 0
	} else if x > dim-1 {
		return dim - 1
	}
	return x
}

type rasterDistances struct {
	table [][]float64
}

func newRasterDistances() *rasterDistances {
	res := &rasterDistances{
		table: make([][]float64, maxNearbyDelta+1),
	}
	for i := range res.table {
		res.table[i] = make([]float64, maxNearbyDelta+1)
		for j := range res.table[i] {
			res.table[i][j] = math.Sqrt(math.Pow(float64(i), 2) + math.Pow(float64(j), 2))
		}
	}
	return res
}

func (r *rasterDistances) Distance(p1, p2 rasterPoint) float64 {
	return r.table[essentials.AbsInt(p1.X-p2.X)][essentials.AbsInt(p1.Y-p2.Y)]
}

type connectorSearchNode struct {
	Point  rasterPoint
	Parent *connectorSearchNode

	NonPreferred        bool
	NumPreferredChanges int
}

func (c *connectorSearchNode) AddStep(p rasterPoint, nonPref bool) *connectorSearchNode {
	changes := c.NumPreferredChanges
	if nonPref != c.NonPreferred {
		changes++
	}
	return &connectorSearchNode{
		Point:  p,
		Parent: c,

		NonPreferred:        nonPref,
		NumPreferredChanges: changes,
	}
}
