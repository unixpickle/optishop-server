package optishop

import (
	"math"

	"github.com/unixpickle/approb"
)

// A TSPSolver is an algorithm that (approximately) solves
// Traveling salesman problems.
type TSPSolver interface {
	// SolveTSP computes a route through n points that is
	// intended to be efficient with respect to the given
	// distance function.
	//
	// The resulting route starts at point 0 and ends at
	// point n - 1. It passes through each point exactly
	// once.
	SolveTSP(n int, distance func(a, b int) float64) []int
}

// GreedyTSPSolver is a TSPSolver that uses the nearest
// neighbor algorithm.
type GreedyTSPSolver struct{}

// SolveTSP generates a greedy solution to the TSP.
func (g GreedyTSPSolver) SolveTSP(n int, distance func(a, b int) float64) []int {
	visited := map[int]bool{0: true}
	route := []int{0}
	p := 0
	for len(route) < n-1 {
		minDist := math.Inf(1)
		minNode := 0
		for i := 1; i < n-1; i++ {
			if visited[i] {
				continue
			}
			dist := distance(p, i)
			if dist < minDist {
				minDist = dist
				minNode = i
			}
		}
		p = minNode
		route = append(route, p)
		visited[p] = true
	}
	route = append(route, n-1)
	return route
}

// FactorialTSPSolver is a TSPSolver that runs in O(n!)
// time but finds exact solutions.
type FactorialTSPSolver struct{}

// SolveTSP generates an exact solution to the TSP in
// O(n!) time.
func (f FactorialTSPSolver) SolveTSP(n int, distance func(a, b int) float64) []int {
	var bestSolution []int
	bestDistance := math.Inf(1)

	for perm := range approb.Perms(n - 2) {
		solution := make([]int, 0, n)
		solution = append(solution, 0)
		for _, x := range perm {
			solution = append(solution, x+1)
		}
		solution = append(solution, n-1)

		var dist float64
		for i := 0; i < len(solution)-1; i++ {
			dist += distance(solution[i], solution[i+1])
		}
		if dist < bestDistance {
			bestDistance = dist
			bestSolution = solution
		}
	}

	return bestSolution
}
