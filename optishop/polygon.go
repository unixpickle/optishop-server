package optishop

import (
	"math"
	"math/rand"
)

var randomDirection Point

func init() {
	theta := rand.Float64() * math.Pi * 2
	randomDirection.X = math.Cos(theta)
	randomDirection.Y = math.Sin(theta)
}

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

// Sub subtracts p1 from p and returns the result.
func (p Point) Sub(p1 Point) Point {
	return Point{X: p.X - p1.X, Y: p.Y - p1.Y}
}

// A Path is a sequence of points leading from some start
// destination to some end destination.
type Path []Point

// Length gets the total path length in Euclidean space.
func (p Path) Length() float64 {
	var res float64
	for i := 1; i < len(p); i++ {
		res += p[i].Distance(p[i-1])
	}
	return res
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

// Contains checks if the polygon contains a given point.
func (p Polygon) Contains(p1 Point) bool {
	p = p.Dedup()
	numIntersections := 0
	r := randomRay(p1)
	for i, start := range p {
		line := &lineSegment{Start: start, End: p.PointAt(i + 1)}
		if rayIntersects(r, line) {
			numIntersections++
		}
	}
	return numIntersections%2 == 1
}

type ray struct {
	Origin    Point
	Direction Point
}

func randomRay(o Point) *ray {
	return &ray{
		Origin:    o,
		Direction: randomDirection,
	}
}

type lineSegment struct {
	Start Point
	End   Point
}

func rayIntersects(r *ray, l *lineSegment) bool {
	endToStart := l.End.Sub(l.Start)
	m11, m12, m21, m22 := r.Direction.X, endToStart.X, r.Direction.Y, endToStart.Y

	// Inverse matrix formula for 2x2 matrices.
	d := 1 / (m11*m22 - m12*m21)
	mInv11, mInv12, mInv21, mInv22 := d*m22, -d*m12, -d*m21, d*m11

	invInput := r.Origin.Sub(l.Start)
	rayExtent := -(mInv11*invInput.X + mInv12*invInput.Y)
	segmentExtent := mInv21*invInput.X + mInv22*invInput.Y

	return rayExtent > 0 && segmentExtent > 0 && segmentExtent < 1
}
