package optishop

import "math"

// A Point is a 2-dimensional location in space.
type Point struct {
	X float64
	Y float64
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
