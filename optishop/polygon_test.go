package optishop

import (
	"math"
	"math/rand"
	"testing"
)

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

func TestConvexPolygonContainment(t *testing.T) {
	poly := ConvexPolygon{
		Point{X: 0, Y: 1},
		Point{X: 1, Y: 0},
		Point{X: 0.5, Y: 0.5},
		Point{X: 4, Y: 1},
		Point{X: 2, Y: 3},
		Point{X: 1, Y: 4},
		Point{X: 4, Y: 6},
	}
	checkPoints := []Point{
		Point{X: 2, Y: 2},
		Point{X: 0.25, Y: 0.25},
		Point{X: 2, Y: 5},
		Point{X: 2, Y: 5},
		Point{X: 2, Y: 4},
	}
	contained := []bool{
		true,
		false,
		false,
		false,
		true,
	}
	for i, p := range checkPoints {
		score := poly.ContainmentScore(p)
		actual := score >= 0
		expected := contained[i]
		if actual != expected {
			t.Errorf("case %d: got %v but expected %v (score %f)", i, actual, expected, score)
		}
	}
}

func TestConvexPolygonContainmentRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		poly := randomConvexPoly()
		for j := 0; j < 100; j++ {
			coeffs := make([]float64, len(poly))
			sum := 0.0
			for k := range coeffs {
				coeffs[k] = rand.Float64()
				sum += coeffs[k]
			}
			combination := Point{}
			for k, p := range poly {
				coeff := coeffs[k] / sum
				combination.X += p.X * coeff
				combination.Y += p.Y * coeff
			}
			if poly.ContainmentScore(combination) < 0 {
				t.Error("convex combination should be contained")
			}
		}
	}
}

func TestConvexPolygonPermutations(t *testing.T) {
	for i := 0; i < 10; i++ {
		poly := randomConvexPoly()
		for j := 0; j < 10; j++ {
			p := Point{X: rand.NormFloat64(), Y: rand.NormFloat64()}
			score := poly.ContainmentScore(p)
			for k := 0; k < 10; k++ {
				rand.Shuffle(len(poly), func(i, j int) {
					poly[i], poly[j] = poly[j], poly[i]
				})
				newScore := poly.ContainmentScore(p)
				if math.Abs(newScore-score) > 1e-8 {
					t.Fatal("not permutation invariant")
				}
			}
		}
	}
}

func TestConvexPolygonConversion(t *testing.T) {
	for i := 0; i < 10; i++ {
		convex := randomConvexPoly()
		poly := convex.Polygon()
		for j := 0; j < 100; j++ {
			p := Point{X: rand.NormFloat64(), Y: rand.NormFloat64()}
			inConvex := convex.ContainmentScore(p) >= 0
			inPoly := NewPolyContainer(poly).Contains(p)
			if inPoly != inConvex {
				t.Fatal("disagreement in containment")
			}
		}
	}
}

func randomConvexPoly() ConvexPolygon {
	res := ConvexPolygon{}
	for i := 0; i < 5+rand.Intn(5); i++ {
		res = append(res, Point{X: rand.NormFloat64(), Y: rand.NormFloat64()})
	}
	for i := 0; i < rand.Intn(3); i++ {
		res = append(res, res[rand.Intn(len(res))])
	}

	// Shuffle so that duplicates aren't necessarily
	// at the end.
	rand.Shuffle(len(res), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})

	return res
}
