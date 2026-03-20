package crawler

import (
	"sort"
	"sync"
)

// PageInfo holds metadata about a crawled page.
type PageInfo struct {
	URL          string
	Depth        int
	StatusCode   int
	IsRedirect   bool
	RedirectTo   string
	InboundCount int
	PageRank     float64
}

// Graph is a thread-safe directed adjacency list.
type Graph struct {
	mu    sync.RWMutex
	edges map[string][]string // from -> []to
	nodes map[string]*PageInfo
	feeds map[string]struct{} // discovered RSS/Atom feed URLs
}

// NewGraph creates an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		edges: make(map[string][]string),
		nodes: make(map[string]*PageInfo),
		feeds: make(map[string]struct{}),
	}
}

// AddDiscoveredFeed records an RSS/Atom feed URL found during crawling.
func (g *Graph) AddDiscoveredFeed(feedURL string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.feeds[feedURL] = struct{}{}
}

// DiscoveredFeeds returns all unique RSS/Atom feed URLs found during the crawl.
func (g *Graph) DiscoveredFeeds() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]string, 0, len(g.feeds))
	for u := range g.feeds {
		out = append(out, u)
	}
	return out
}

// AddNode upserts a PageInfo. Existing entries are not overwritten.
func (g *Graph) AddNode(info *PageInfo) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.nodes[info.URL]; !exists {
		g.nodes[info.URL] = info
	}
}

// AddEdge records a directed link from → to, ensuring both nodes exist.
func (g *Graph) AddEdge(from, to string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.nodes[from]; !exists {
		g.nodes[from] = &PageInfo{URL: from}
	}
	if _, exists := g.nodes[to]; !exists {
		g.nodes[to] = &PageInfo{URL: to}
	}
	g.edges[from] = append(g.edges[from], to)
}

// Nodes returns a snapshot of all nodes (read-only copy of pointers).
func (g *Graph) Nodes() map[string]*PageInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[string]*PageInfo, len(g.nodes))
	for k, v := range g.nodes {
		out[k] = v
	}
	return out
}

// Edges returns a snapshot of the adjacency list.
func (g *Graph) Edges() map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[string][]string, len(g.edges))
	for k, v := range g.edges {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// NodeCount returns the number of nodes.
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// ComputeInboundCounts tallies how many edges point to each node.
func (g *Graph) ComputeInboundCounts() {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Reset
	for _, n := range g.nodes {
		n.InboundCount = 0
	}
	for _, targets := range g.edges {
		for _, t := range targets {
			if n, ok := g.nodes[t]; ok {
				n.InboundCount++
			}
		}
	}
}

// OrphanedPages returns pages with no inbound links (excluding the seed/root at depth 0).
func (g *Graph) OrphanedPages() []*PageInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []*PageInfo
	for _, n := range g.nodes {
		if n.InboundCount == 0 && n.Depth > 0 {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out
}

// BrokenLinks returns pages that returned HTTP 404.
func (g *Graph) BrokenLinks() []*PageInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []*PageInfo
	for _, n := range g.nodes {
		if n.StatusCode == 404 {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out
}

// TopByInbound returns the top n pages sorted by InboundCount descending.
func (g *Graph) TopByInbound(n int) []*PageInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	all := make([]*PageInfo, 0, len(g.nodes))
	for _, p := range g.nodes {
		all = append(all, p)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].InboundCount > all[j].InboundCount
	})
	if n > 0 && n < len(all) {
		return all[:n]
	}
	return all
}

// TopByPageRank returns the top n pages sorted by PageRank descending.
func (g *Graph) TopByPageRank(n int) []*PageInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	all := make([]*PageInfo, 0, len(g.nodes))
	for _, p := range g.nodes {
		all = append(all, p)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].PageRank > all[j].PageRank
	})
	if n > 0 && n < len(all) {
		return all[:n]
	}
	return all
}

// RedirectChains returns chains of 3 or more hops.
// Each chain is a slice of URLs starting from the initiating URL.
func (g *Graph) RedirectChains() [][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var chains [][]string
	for _, n := range g.nodes {
		if !n.IsRedirect {
			continue
		}
		chain := []string{n.URL}
		current := n.RedirectTo
		for hops := 0; hops < 10 && current != ""; hops++ {
			chain = append(chain, current)
			next, ok := g.nodes[current]
			if !ok || !next.IsRedirect {
				break
			}
			current = next.RedirectTo
		}
		if len(chain) >= 3 {
			chains = append(chains, chain)
		}
	}
	return chains
}
