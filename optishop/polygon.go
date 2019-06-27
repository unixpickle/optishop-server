package optishop

import "math"

// A Point is a 2-dimensional location in space.
type Point struct {
	X float64
	Y float64
}

// Midpoint computes the point directly between p1 and p2.
func Midpoint(p1, p2 Point) Point {
	return Point{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
}

// Distance computes the Euclidean distance from p to p1.
func (p Point) Distance(p1 Point) float64 {
	return math.Sqrt(math.Pow(p.X-p1.X, 2) + math.Pow(p.Y-p1.Y, 2))
}

// A Polygon is an arbitrary closed path.
// It is obtained by tracing a path from the first point
// to the last, and then back to the first point again.
type Polygon []Point

// Bounds computes the bounding box of the polygon.
func (p Polygon) Bounds() (x, y, width, height float64) {
	min := p[0]
	max := p[0]
	for _, point := range p {
		min.X = math.Min(min.X, point.X)
		min.Y = math.Min(min.Y, point.Y)
		max.X = math.Max(max.X, point.X)
		max.Y = math.Max(max.Y, point.Y)
	}
	return min.X, min.Y, max.X - min.X, max.Y - min.Y
}

// Dedup creates a Polygon that definitely does not
// contain identical start and end points.
func (p Polygon) Dedup() Polygon {
	if p[0] == p[len(p)-1] {
		return p[:len(p)-1]
	}
	return p
}

// PointAt gets the point at the given index, but it
// allows the index to be negative or greater than the
// number of points (in which case wrapping is done).
func (p Polygon) PointAt(idx int) Point {
	for idx < 0 {
		idx += len(p)
	}
	return p[idx%len(p)]
}
