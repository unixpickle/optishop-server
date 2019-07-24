package optishop

import (
	"math"
	"sort"
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

// SolveTSP solves a Traveling salesman problem quickly,
// potentially using an approximation if an exact solution
// is infeasible.
//
// This function is like TSPSolver.SolveTSP, except that
// it automatically selects an appropriate TSPSolver.
func SolveTSP(n int, distance func(a, b int) float64) []int {
	if n <= 10 {
		return (FactorialTSPSolver{}).SolveTSP(n, distance)
	} else if n <= 30 {
		return (BeamTSPSolver{BeamSize: 1000}).SolveTSP(n, distance)
	} else if n <= 50 {
		return (BeamTSPSolver{BeamSize: 100}).SolveTSP(n, distance)
	} else {
		return (GreedyTSPSolver{}).SolveTSP(n, distance)
	}
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

// BeamTSPSolver is a TSPSolver that uses beam search to
// find good solutions which are optimal if the search
// problem is small, but always fast regardless of the
// size of the search problem.
type BeamTSPSolver struct {
	// BeamSize is the number of search nodes to keep
	// around at every step of the algorithm.
	BeamSize int
}

// SolveTSP generates an approximate (or sometimes excact)
// solution to the TSP in no more than O(n^3 * BeamSize).
func (b BeamTSPSolver) SolveTSP(n int, distance func(a, b int) float64) []int {
	nodes := []beamSearchNode{beamSearchNode{solution: []int{0}}}

	for i := 0; i < n-2; i++ {
		newNodes := make([]beamSearchNode, 0, len(nodes)*(n-2-i))
		for _, node := range nodes {
			for j := 1; j < n-1; j++ {
				if !node.containsStop(j) {
					newNodes = append(newNodes, beamSearchNode{
						solution: append(append([]int{}, node.solution...), j),
						distance: node.distance + distance(node.solution[i], j),
					})
				}
			}
		}
		nodes = newNodes
		if len(nodes) > b.BeamSize {
			sort.Slice(nodes, func(i, j int) bool {
				return nodes[i].distance < nodes[j].distance
			})
			nodes = nodes[:b.BeamSize]
		}
	}

	for i := range nodes {
		node := &nodes[i]
		node.distance += distance(node.solution[len(node.solution)-1], n-1)
		node.solution = append(node.solution, n-1)
	}

	shortest := nodes[0].distance
	solution := nodes[0].solution

	for _, node := range nodes {
		if node.distance < shortest {
			shortest = node.distance
			solution = node.solution
		}
	}

	return solution
}

type beamSearchNode struct {
	solution []int
	distance float64
}

func (b *beamSearchNode) containsStop(idx int) bool {
	for _, x := range b.solution {
		if x == idx {
			return true
		}
	}
	return false
}
