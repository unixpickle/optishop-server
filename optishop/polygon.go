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

type lineSegment struct {
	Start Point
	End   Point
}

// A PolyContainer can check if a polygon contains any
// arbitrary point.
type PolyContainer struct {
	intersectors []*rayIntersector
}

// NewPolyContainer creates a PolyContainer for the given
// polygon.
// This performs a lot of up-front computation that can be
// amortized over the course of many containment checks.
func NewPolyContainer(poly Polygon) *PolyContainer {
	p := poly.Dedup()
	res := &PolyContainer{}
	for i, start := range p {
		line := &lineSegment{Start: start, End: p.PointAt(i + 1)}
		res.intersectors = append(res.intersectors, newRayIntersector(randomDirection, line))
	}
	return res
}

// Contains checks if the polygon contains a point.
func (p *PolyContainer) Contains(point Point) bool {
	numIntersections := 0
	for _, intersector := range p.intersectors {
		if intersector.Intersects(point) {
			numIntersections++
		}
	}
	return numIntersections%2 == 1
}

// A rayIntersector checks if rays intersect a line
// segment.
// A rayIntersector is initialized with the direction of
// the ray and the line segment to check, and then can
// quickly check intersections for arbitrary ray starting
// points.
type rayIntersector struct {
	mInv11 float64
	mInv12 float64
	mInv21 float64
	mInv22 float64

	start Point
}

func newRayIntersector(direction Point, l *lineSegment) *rayIntersector {
	endToStart := l.End.Sub(l.Start)
	m11, m12, m21, m22 := direction.X, endToStart.X, direction.Y, endToStart.Y

	// Inverse matrix formula for 2x2 matrices.
	d := 1 / (m11*m22 - m12*m21)
	return &rayIntersector{
		mInv11: d * m22,
		mInv12: -d * m12,
		mInv21: -d * m21,
		mInv22: d * m11,
		start:  l.Start,
	}
}

func (r *rayIntersector) Intersects(origin Point) bool {
	invInput := origin.Sub(r.start)
	rayExtent := -(r.mInv11*invInput.X + r.mInv12*invInput.Y)
	segmentExtent := r.mInv21*invInput.X + r.mInv22*invInput.Y
	return rayExtent > 0 && segmentExtent > 0 && segmentExtent < 1
}
