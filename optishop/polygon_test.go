package optishop

import "testing"

type ray struct {
	Origin    Point
	Direction Point
}

func TestRayIntersects(t *testing.T) {
	rays := []*ray{
		&ray{
			Origin:    Point{X: 1, Y: 2},
			Direction: Point{X: 2, Y: 3},
		},
		&ray{
			Origin:    Point{X: 1, Y: 2},
			Direction: Point{X: 2, Y: 3},
		},
		&ray{
			Origin:    Point{X: 1, Y: 2},
			Direction: Point{X: -2, Y: -3},
		},
		&ray{
			Origin:    Point{X: 1, Y: 2},
			Direction: Point{X: 2, Y: 3},
		},
	}
	segments := []*lineSegment{
		&lineSegment{
			Start: Point{X: 4, Y: 5},
			End:   Point{X: 3, Y: 6},
		},
		&lineSegment{
			Start: Point{X: 3, Y: 6},
			End:   Point{X: 4, Y: 5},
		},
		&lineSegment{
			Start: Point{X: 4, Y: 5},
			End:   Point{X: 3, Y: 6},
		},
		&lineSegment{
			Start: Point{X: 4, Y: 5},
			End:   Point{X: 5, Y: 4},
		},
	}
	expected := []bool{
		true,
		true,
		false,
		false,
	}
	for i, exp := range expected {
		actual := newRayIntersector(rays[i].Direction, segments[i]).Intersects(rays[i].Origin)
		if actual != exp {
			t.Errorf("case %d: expected %v but got %v", i, exp, actual)
		}
	}
}

func TestPolygonContains(t *testing.T) {
	poly := Polygon{
		Point{X: 0, Y: 0},
		Point{X: 2, Y: 1},
		Point{X: 3, Y: 4},
		Point{X: 1, Y: 3},
		Point{X: 0, Y: 5},
		Point{X: -4, Y: -2},
		Point{X: 0, Y: 1},
		Point{X: 0, Y: 0},
	}
	points := []Point{
		Point{X: 0, Y: 2},
		Point{X: 0, Y: -1},
		Point{X: -2, Y: 0},
		Point{X: -2, Y: 1},
		Point{X: -2, Y: 2},
		Point{X: -1, Y: 0},
		Point{X: 2, Y: 3},
		Point{X: 1, Y: 4},
	}
	contained := []bool{
		true,
		false,
		true,
		true,
		false,
		false,
		true,
		false,
	}
	container := NewPolyContainer(poly)
	for i, expected := range contained {
		actual := container.Contains(points[i])
		if actual != expected {
			t.Errorf("case %d: expected %v but got %v", i, expected, actual)
		}
	}
}
