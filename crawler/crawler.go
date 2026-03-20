package crawler

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Config controls crawl behaviour.
type Config struct {
	SeedURL   string
	MaxDepth  int
	MaxPages  int
	Workers   int
	RateLimit rate.Limit    // requests/second
	Timeout   time.Duration // per-request timeout
}

func (c *Config) setDefaults() {
	if c.Workers <= 0 {
		c.Workers = 10
	}
	if c.RateLimit <= 0 {
		c.RateLimit = 2.0
	}
	if c.Timeout <= 0 {
		c.Timeout = 10 * time.Second
	}
	if c.MaxDepth <= 0 {
		c.MaxDepth = 3
	}
	if c.MaxPages <= 0 {
		c.MaxPages = 500
	}
}

type job struct {
	url   string
	depth int
	from  string
}

// Engine performs a BFS crawl of a domain using a worker pool.
type Engine struct {
	cfg     Config
	graph   *Graph
	client  *http.Client
	limiter *rate.Limiter
	visited sync.Map
	queue   chan job
	wg      sync.WaitGroup
	count   atomic.Int64
}

// NewEngine creates a ready-to-run Engine.
func NewEngine(cfg Config) *Engine {
	cfg.setDefaults()
	return &Engine{
		cfg:   cfg,
		graph: NewGraph(),
		client: &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		limiter: rate.NewLimiter(cfg.RateLimit, 1),
	}
}

// Run crawls until the context is cancelled, max pages/depth reached, or the queue drains.
// Returns the populated Graph.
func (e *Engine) Run(ctx context.Context) (*Graph, error) {
	e.queue = make(chan job, e.cfg.Workers*4)

	// Start workers
	for i := 0; i < e.cfg.Workers; i++ {
		go e.worker(ctx)
	}

	// Seed
	e.wg.Add(1)
	e.queue <- job{url: e.cfg.SeedURL, depth: 0}

	// Wait for all work to finish then close the queue
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// drain workers
		close(e.queue)
		return e.graph, nil
	case <-ctx.Done():
		close(e.queue)
		return e.graph, ctx.Err()
	}
}

func (e *Engine) worker(ctx context.Context) {
	for j := range e.queue {
		e.processJob(ctx, j)
	}
}

// PageCount returns the number of pages fetched so far. Safe to call concurrently.
func (e *Engine) PageCount() int64 {
	return e.count.Load()
}

func (e *Engine) processJob(ctx context.Context, j job) {
	defer e.wg.Done()

	if ctx.Err() != nil {
		return
	}

	// Hard stop: jobs already queued before the limit was reached must not run
	if e.count.Load() >= int64(e.cfg.MaxPages) {
		return
	}

	// Rate limit
	if err := e.limiter.Wait(ctx); err != nil {
		return
	}

	// Per-request timeout
	reqCtx, cancel := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, j.url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "WebCrawler/1.0")

	resp, err := e.client.Do(req)
	if err != nil {
		// Still record the node as errored
		e.graph.AddNode(&PageInfo{URL: j.url, Depth: j.depth, StatusCode: 0})
		if j.from != "" {
			e.graph.AddEdge(j.from, j.url)
		}
		return
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	isRedirect := finalURL != j.url
	redirectTo := ""
	if isRedirect {
		redirectTo = finalURL
	}

	info := &PageInfo{
		URL:        j.url,
		Depth:      j.depth,
		StatusCode: resp.StatusCode,
		IsRedirect: isRedirect,
		RedirectTo: redirectTo,
	}
	e.graph.AddNode(info)
	if j.from != "" {
		e.graph.AddEdge(j.from, j.url)
	}

	e.count.Add(1)

	// Only extract links from successful HTML responses
	if resp.StatusCode != http.StatusOK || j.depth >= e.cfg.MaxDepth {
		return
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "text/html") {
		return
	}

	base, err := url.Parse(j.url)
	if err != nil {
		return
	}

	extracted := Extract(base, resp.Body)
	for _, feedURL := range extracted.Feeds {
		e.graph.AddDiscoveredFeed(feedURL)
	}
	for _, link := range extracted.Links {
		if e.count.Load() >= int64(e.cfg.MaxPages) {
			break
		}
		if _, loaded := e.visited.LoadOrStore(link, true); loaded {
			continue
		}
		e.wg.Add(1)
		select {
		case e.queue <- job{url: link, depth: j.depth + 1, from: j.url}:
		default:
			// Queue full — drop this link
			e.wg.Done()
		}
	}
}

// Graph returns the engine's graph (safe to call after Run returns).
func (e *Engine) Graph() *Graph {
	return e.graph
}

// RunFull crawls, then computes inbound counts and PageRank on the result.
// Call PageCount() concurrently to read live progress while this is running.
func (e *Engine) RunFull(ctx context.Context) (*Graph, error) {
	e.visited.LoadOrStore(e.cfg.SeedURL, true)
	g, err := e.Run(ctx)
	if err != nil && err != context.Canceled {
		return g, err
	}
	g.ComputeInboundCounts()
	ComputePageRank(g, 0.85, 100)
	return g, nil
}

// RunCrawl is kept for backwards compatibility.
func RunCrawl(ctx context.Context, cfg Config) (*Graph, error) {
	return NewEngine(cfg).RunFull(ctx)
}
