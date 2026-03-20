package crawler

import "math"

// ComputePageRank runs iterative PageRank on g.
// damping is typically 0.85, maxIter caps iterations.
// Convergence threshold is 1e-6.
func ComputePageRank(g *Graph, damping float64, maxIter int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	n := len(g.nodes)
	if n == 0 {
		return
	}

	// Collect URLs for stable iteration
	urls := make([]string, 0, n)
	for u := range g.nodes {
		urls = append(urls, u)
	}

	// Initial rank
	init := 1.0 / float64(n)
	rank := make(map[string]float64, n)
	for _, u := range urls {
		rank[u] = init
	}

	// Out-degree map
	outDeg := make(map[string]int, n)
	for u, targets := range g.edges {
		outDeg[u] = len(targets)
	}

	// Build reverse edge map: to -> []from
	inbound := make(map[string][]string, n)
	for u, targets := range g.edges {
		for _, t := range targets {
			inbound[t] = append(inbound[t], u)
		}
	}

	for iter := 0; iter < maxIter; iter++ {
		// Dangling node mass (nodes with no outgoing links)
		danglingSum := 0.0
		for _, u := range urls {
			if outDeg[u] == 0 {
				danglingSum += rank[u]
			}
		}

		next := make(map[string]float64, n)
		for _, u := range urls {
			// Base teleport + dangling redistribution
			base := (1-damping)/float64(n) + damping*danglingSum/float64(n)
			sum := 0.0
			for _, from := range inbound[u] {
				if od := outDeg[from]; od > 0 {
					sum += rank[from] / float64(od)
				}
			}
			next[u] = base + damping*sum
		}

		// Check convergence
		delta := 0.0
		for _, u := range urls {
			delta += math.Abs(next[u] - rank[u])
		}
		rank = next
		if delta < 1e-6 {
			break
		}
	}

	for _, u := range urls {
		g.nodes[u].PageRank = rank[u]
	}
}
