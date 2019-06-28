package optishop

import "math"

// A Path is a sequence of points leading from some start
// destination to some end destination.
type Path []Point

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
	// TODO: dynamic way to compute an ideal raster size.
	return newRasterConnector(f, 600, 600)
}

type rasterConnector struct {
	boundsX      float64
	boundsY      float64
	boundsWidth  float64
	boundsHeight float64

	width  int
	height int

	obstructed []bool
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

		obstructed: make([]bool, width*height),
	}
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := res.xyToPoint(x, y)
			if !floor.Bounds.Contains(p) {
				res.obstructed[idx] = true
			} else {
				for _, poly := range floor.Obstacles {
					if poly.Contains(p) {
						res.obstructed[idx] = true
						break
					}
				}
			}
			idx++
		}
	}
	return res
}

func (r *rasterConnector) Obstructed(p Point) bool {
	x, y := r.pointToXY(p)
	if x < 0 || y < 0 || x >= r.width || y >= r.height {
		return true
	}
	return r.obstructed[r.xyToIndex(x, y)]
}

func (r *rasterConnector) Unobstruct(p Point) Point {
	// Basic BFS to find the nearest point in L1 distance.
	x, y := r.pointToXY(p)
	queue := [][2]int{[2]int{x, y}}
	visited := map[[2]int]bool{queue[0]: true}
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		if !r.obstructed[r.xyToIndex(next[0], next[1])] {
			return r.xyToPoint(next[0], next[1])
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
	x, y := r.pointToXY(r.Unobstruct(a))
	start := [2]int{x, y}
	x, y = r.pointToXY(r.Unobstruct(b))
	end := [2]int{x, y}

	// TODO: use a priority queue here so that we can use
	// Euclidean distance instead of L1 distance.
	queue := [][][2]int{{start}}
	visited := map[[2]int]bool{start: true}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node[len(node)-1] == end {
			points := Path{a}
			for _, xy := range node {
				points = append(points, r.xyToPoint(xy[0], xy[1]))
			}
			points = append(points, b)
			return points
		}
		for _, newPoint := range r.neighbors(node[len(node)-1]) {
			if !visited[newPoint] && !r.obstructed[r.xyToIndex(newPoint[0], newPoint[1])] {
				visited[newPoint] = true
				newNode := append(append([][2]int{}, node...), newPoint)
				queue = append(queue, newNode)
			}
		}
	}

	return nil
}

func (r *rasterConnector) pointToXY(p Point) (int, int) {
	x := int(math.Round(float64(r.width) * (p.X - r.boundsX) / r.boundsWidth))
	y := int(math.Round(float64(r.height) * (p.Y - r.boundsY) / r.boundsHeight))
	return x, y
}

func (r *rasterConnector) xyToPoint(x, y int) Point {
	return Point{
		X: (float64(x)/float64(r.width))*r.boundsWidth + r.boundsX,
		Y: (float64(y)/float64(r.height))*r.boundsHeight + r.boundsY,
	}
}

func (r *rasterConnector) xyToIndex(x, y int) int {
	return x + y*r.width
}

func (r *rasterConnector) neighbors(xy [2]int) [][2]int {
	res := make([][2]int, 0, 4)
	x, y := xy[0], xy[1]
	if x > 0 {
		res = append(res, [2]int{x - 1, y})
	}
	if y > 0 {
		res = append(res, [2]int{x, y - 1})
	}
	if x+1 < r.width {
		res = append(res, [2]int{x + 1, y})
	}
	if y+1 < r.height {
		res = append(res, [2]int{x, y + 1})
	}
	return res
}
