package optishop

import (
	"fmt"
	"testing"
)

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

func TestBeamTSPSolver(t *testing.T) {
	distances := testingTSPProblem()
	solution := (BeamTSPSolver{BeamSize: 100000}).SolveTSP(10, distances)
	expected := []int{0, 8, 3, 1, 2, 7, 6, 5, 4, 9}
	fmt.Println(solution)
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

func BenchmarkBeamTSPSolver(b *testing.B) {
	distances := testingTSPProblem()
	b.Run("Beam100K", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			(BeamTSPSolver{BeamSize: 100000}).SolveTSP(10, distances)
		}
	})
	b.Run("Beam1K", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			(BeamTSPSolver{BeamSize: 1000}).SolveTSP(10, distances)
		}
	})
	b.Run("Beam1KEntries30", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			(BeamTSPSolver{BeamSize: 1000}).SolveTSP(30, distances)
		}
	})
	b.Run("Beam100Entries50", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			(BeamTSPSolver{BeamSize: 100}).SolveTSP(50, distances)
		}
	})
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
		{0.115060, 0.199443},
		{0.576624, 0.587014},
		{0.205503, 0.684167},
		{0.078758, 0.770373},
		{0.555492, 0.061625},
		{0.619535, 0.193413},
		{0.675853, 0.018615},
		{0.003038, 0.360655},
		{0.943927, 0.090063},
		{0.136943, 0.873513},
		{0.596813, 0.256981},
		{0.028375, 0.739079},
		{0.569179, 0.855967},
		{0.590624, 0.027340},
		{0.855540, 0.335992},
		{0.256435, 0.595910},
		{0.011180, 0.640155},
		{0.200203, 0.741108},
		{0.742390, 0.724186},
		{0.038512, 0.649543},
		{0.634280, 0.968998},
		{0.370886, 0.574291},
		{0.309292, 0.210056},
		{0.975317, 0.204628},
		{0.582947, 0.203376},
		{0.837466, 0.100211},
		{0.547003, 0.445724},
		{0.045748, 0.610894},
		{0.704890, 0.101921},
		{0.572273, 0.636992},
		{0.003491, 0.423774},
		{0.304923, 0.557073},
		{0.411545, 0.197108},
		{0.003603, 0.297628},
		{0.391232, 0.207880},
		{0.330168, 0.667484},
		{0.534692, 0.658085},
		{0.564341, 0.850036},
		{0.197909, 0.325590},
		{0.696856, 0.976566},
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
