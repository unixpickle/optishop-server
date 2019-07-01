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

	res.checkBoundaries(floor.Bounds)
	res.addObstacles(floor.Obstacles)

	return res
}

func (r *rasterConnector) Obstructed(p Point) bool {
	rp := r.pointToRaster(p)
	if rp.X < 0 || rp.Y < 0 || rp.X >= r.width || rp.Y >= r.height {
		return true
	}
	return r.obstructed[r.pointToIndex(rp)]
}

func (r *rasterConnector) Unobstruct(p Point) Point {
	// Basic BFS to find the nearest point in L1 distance.
	start := r.pointToRaster(p)
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

	// TODO: use a priority queue here so that we can use
	// Euclidean distance instead of L1 distance.
	queue := [][]rasterPoint{{start}}
	visited := map[rasterPoint]bool{start: true}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node[len(node)-1] == end {
			points := Path{a}
			for _, xy := range node {
				points = append(points, r.rasterToPoint(xy))
			}
			points = append(points, b)
			return points
		}
		for _, newPoint := range r.neighbors(node[len(node)-1]) {
			if !visited[newPoint] && !r.obstructed[r.pointToIndex(newPoint)] {
				visited[newPoint] = true
				newNode := append(append([]rasterPoint{}, node...), newPoint)
				queue = append(queue, newNode)
			}
		}
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

func (r *rasterConnector) addObstacles(polygons []Polygon) {
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
					r.obstructed[r.pointToIndex(rp)] = true
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

func clampDim(x, dim int) int {
	if x < 0 {
		return 0
	} else if x > dim-1 {
		return dim - 1
	}
	return x
}
