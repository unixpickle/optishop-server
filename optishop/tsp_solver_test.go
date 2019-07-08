package optishop

import "testing"

func TestFactorialTSPSolver(t *testing.T) {
	distances := testingTSPProblem()
	solution := (FactorialTSPSolver{}).SolveTSP(10, distances)
	expected := []int{0, 8, 3, 1, 2, 7, 6, 5, 4, 9}
	if len(solution) != len(expected) {
		t.Fatal("incorrect length")
	}
	for i, x := range expected {
		if x != solution[i] {
			t.Errorf("expected solution %v but got %v", expected, solution)
			break
		}
	}
}

func BenchmarkFactorialTSPSolver(b *testing.B) {
	distances := testingTSPProblem()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		(FactorialTSPSolver{}).SolveTSP(10, distances)
	}
}

func testingTSPProblem() func(i, j int) float64 {
	// Points were randomly generated uniformly in [0, 1].
	points := []Point{
		{0.322865, 0.944045},
		{0.990716, 0.845700},
		{0.973359, 0.696303},
		{0.884750, 0.872275},
		{0.347809, 0.669670},
		{0.432515, 0.595936},
		{0.437946, 0.434378},
		{0.966726, 0.511736},
		{0.459672, 0.786280},
		{0.165782, 0.405137},
	}
	distances := make([][]float64, len(points))
	for i, p := range points {
		distances[i] = make([]float64, len(points))
		for j, p1 := range points {
			distances[i][j] = p.Distance(p1)
		}
	}
	return func(i, j int) float64 {
		return distances[i][j]
	}
}
