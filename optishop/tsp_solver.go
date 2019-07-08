package optishop

import (
	"math"
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
	solution := make([]int, n)
	solution[n-1] = n - 1
	bestDist := math.Inf(1)

	used := make([]bool, n-1)
	used[0] = true

	f.recurse(1, 0, used, make([]int, n-1), distance, &bestDist, solution)

	return solution
}

func (f FactorialTSPSolver) recurse(length int, distance float64, used []bool, perm []int,
	distFn func(i, j int) float64, bestDist *float64, bestPerm []int) {
	if length == len(used) {
		distance += distFn(perm[length-1], len(used))
		if distance < *bestDist {
			*bestDist = distance
			copy(bestPerm, perm)
		}
		return
	}
	for i, u := range used {
		if !u {
			perm[length] = i
			used[i] = true
			f.recurse(length+1, distance+distFn(perm[length-1], i), used, perm, distFn,
				bestDist, bestPerm)
			used[i] = false
		}
	}
}
