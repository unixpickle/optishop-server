package optishop

import "math"

// A TSPSolver is an algorithm that (approximately) solves
// Travelling salesman problems.
type TSPSolver interface {
	// SolveTSP computes a route through n points that is
	// intended to be efficient with respect to the given
	// distance function.
	//
	// The resulting route both starts and ends at the
	// first point, and passes through the other points
	// exactly once.
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
	for len(route) < n {
		minDist := math.Inf(1)
		minNode := 0
		for i := 1; i < n; i++ {
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
	route = append(route, 0)
	return route
}
