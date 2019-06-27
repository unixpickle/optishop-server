package optishop

// A Polygon is an arbitrary closed path.
// It is obtained by tracing a path from the first point
// to the last, and then back to the first point again.
type Polygon []Point

// A Point is a 2-dimensional location in space.
type Point struct {
	X float64
	Y float64
}
