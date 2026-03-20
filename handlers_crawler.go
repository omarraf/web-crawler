package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/omarraf/web-scraper/crawler"
	"github.com/omarraf/web-scraper/internal/database"
)

// POST /v1/crawl_jobs
func (apiCfg *apiConfig) handlerCreateCrawlJob(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		SeedURL  string `json:"seed_url"`
		MaxDepth int    `json:"max_depth"`
		MaxPages int    `json:"max_pages"`
	}
	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "Invalid request body")
		return
	}
	if params.SeedURL == "" {
		respondWithError(w, 400, "seed_url is required")
		return
	}
	if params.MaxDepth <= 0 {
		params.MaxDepth = 3
	}
	if params.MaxPages <= 0 {
		params.MaxPages = 500
	}

	job, err := apiCfg.DB.CreateCrawlJob(r.Context(), database.CreateCrawlJobParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		SeedUrl:   params.SeedURL,
		MaxDepth:  int32(params.MaxDepth),
		MaxPages:  int32(params.MaxPages),
	})
	if err != nil {
		respondWithError(w, 500, "Failed to create crawl job")
		return
	}

	go apiCfg.runCrawlJob(job.ID, crawler.Config{
		SeedURL:  params.SeedURL,
		MaxDepth: params.MaxDepth,
		MaxPages: params.MaxPages,
	})

	respondWithJSON(w, 201, databaseCrawlJobToCrawlJob(job))
}

// GET /v1/crawl_jobs
func (apiCfg *apiConfig) handlerListCrawlJobs(w http.ResponseWriter, r *http.Request, user database.User) {
	jobs, err := apiCfg.DB.GetCrawlJobsByUserID(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 500, "Failed to list crawl jobs")
		return
	}
	out := make([]CrawlJob, len(jobs))
	for i, j := range jobs {
		out[i] = databaseCrawlJobToCrawlJob(j)
	}
	respondWithJSON(w, 200, out)
}

// GET /v1/crawl_jobs/{jobID}
func (apiCfg *apiConfig) handlerGetCrawlJob(w http.ResponseWriter, r *http.Request, user database.User) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		respondWithError(w, 400, "Invalid job ID")
		return
	}
	job, err := apiCfg.DB.GetCrawlJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 404, "Crawl job not found")
		return
	}
	if job.UserID != user.ID {
		respondWithError(w, 403, "Forbidden")
		return
	}
	respondWithJSON(w, 200, databaseCrawlJobToCrawlJob(job))
}

// DELETE /v1/crawl_jobs/{jobID}
func (apiCfg *apiConfig) handlerDeleteCrawlJob(w http.ResponseWriter, r *http.Request, user database.User) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		respondWithError(w, 400, "Invalid job ID")
		return
	}
	job, err := apiCfg.DB.GetCrawlJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 404, "Crawl job not found")
		return
	}
	if job.UserID != user.ID {
		respondWithError(w, 403, "Forbidden")
		return
	}
	if cancel, ok := apiCfg.CrawlEngines.Load(jobID); ok {
		cancel.(context.CancelFunc)()
	}
	respondWithJSON(w, 200, struct{}{})
}

// GET /v1/crawl_jobs/{jobID}/graph
func (apiCfg *apiConfig) handlerGetGraph(w http.ResponseWriter, r *http.Request, user database.User) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		respondWithError(w, 400, "Invalid job ID")
		return
	}
	job, err := apiCfg.DB.GetCrawlJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 404, "Crawl job not found")
		return
	}
	if job.UserID != user.ID {
		respondWithError(w, 403, "Forbidden")
		return
	}

	pages, err := apiCfg.DB.GetPagesByCrawlJob(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch pages")
		return
	}
	links, err := apiCfg.DB.GetPageLinksByCrawlJob(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch links")
		return
	}

	type node struct {
		ID           string  `json:"id"`
		Depth        int32   `json:"depth"`
		PageRank     float64 `json:"page_rank"`
		InboundCount int32   `json:"inbound_count"`
		StatusCode   int32   `json:"status_code,omitempty"`
		IsRedirect   bool    `json:"is_redirect,omitempty"`
	}
	type edge struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	nodes := make([]node, len(pages))
	for i, p := range pages {
		sc := int32(0)
		if p.StatusCode.Valid {
			sc = p.StatusCode.Int32
		}
		nodes[i] = node{
			ID:           p.Url,
			Depth:        p.Depth,
			PageRank:     p.PageRank,
			InboundCount: p.InboundCount,
			StatusCode:   sc,
			IsRedirect:   p.IsRedirect,
		}
	}
	edges := make([]edge, len(links))
	for i, l := range links {
		edges[i] = edge{Source: l.FromUrl, Target: l.ToUrl}
	}

	respondWithJSON(w, 200, map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	})
}

// GET /v1/crawl_jobs/{jobID}/analysis
func (apiCfg *apiConfig) handlerGetAnalysis(w http.ResponseWriter, r *http.Request, user database.User) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		respondWithError(w, 400, "Invalid job ID")
		return
	}
	job, err := apiCfg.DB.GetCrawlJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 404, "Crawl job not found")
		return
	}
	if job.UserID != user.ID {
		respondWithError(w, 403, "Forbidden")
		return
	}

	pages, err := apiCfg.DB.GetPagesByCrawlJob(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch pages")
		return
	}
	links, err := apiCfg.DB.GetPageLinksByCrawlJob(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch links")
		return
	}

	type pageEntry struct {
		URL          string  `json:"url"`
		Depth        int32   `json:"depth"`
		StatusCode   int32   `json:"status_code,omitempty"`
		InboundCount int32   `json:"inbound_count"`
		PageRank     float64 `json:"page_rank"`
		RedirectTo   string  `json:"redirect_to,omitempty"`
	}

	var orphans, broken, authority []pageEntry
	redirectMap := make(map[string]string)
	var redirectPages []pageEntry
	depthMap := make(map[int32]int)

	for _, p := range pages {
		sc := int32(0)
		if p.StatusCode.Valid {
			sc = p.StatusCode.Int32
		}
		rt := ""
		if p.RedirectTo.Valid {
			rt = p.RedirectTo.String
			redirectMap[p.Url] = rt
		}
		entry := pageEntry{URL: p.Url, Depth: p.Depth, StatusCode: sc, InboundCount: p.InboundCount, PageRank: p.PageRank, RedirectTo: rt}
		if p.InboundCount == 0 && p.Depth > 0 {
			orphans = append(orphans, entry)
		}
		if sc == 404 {
			broken = append(broken, entry)
		}
		if p.InboundCount > 0 {
			authority = append(authority, entry)
		}
		if p.IsRedirect {
			redirectPages = append(redirectPages, entry)
		}
		depthMap[p.Depth]++
	}

	// Sort authority by inbound count desc
	for i := 0; i < len(authority); i++ {
		for j := i + 1; j < len(authority); j++ {
			if authority[j].InboundCount > authority[i].InboundCount {
				authority[i], authority[j] = authority[j], authority[i]
			}
		}
	}
	if len(authority) > 20 {
		authority = authority[:20]
	}

	type redirectChain struct {
		Chain []string `json:"chain"`
		Hops  int      `json:"hops"`
	}
	var chains []redirectChain
	for _, rp := range redirectPages {
		chain := []string{rp.URL}
		cur := rp.RedirectTo
		for hops := 0; hops < 10 && cur != ""; hops++ {
			chain = append(chain, cur)
			if next, ok := redirectMap[cur]; ok {
				cur = next
			} else {
				break
			}
		}
		if len(chain) >= 3 {
			chains = append(chains, redirectChain{Chain: chain, Hops: len(chain) - 1})
		}
	}

	// Discovered RSS/Atom feeds stored as newline-separated list
	var discoveredFeeds []string
	if job.DiscoveredFeeds != "" {
		for _, f := range strings.Split(job.DiscoveredFeeds, "\n") {
			if f = strings.TrimSpace(f); f != "" {
				discoveredFeeds = append(discoveredFeeds, f)
			}
		}
	}

	respondWithJSON(w, 200, map[string]interface{}{
		"orphaned_pages":   orphans,
		"broken_links":     broken,
		"top_authority":    authority,
		"redirect_chains":  chains,
		"depth_map":        depthMap,
		"total_pages":      len(pages),
		"total_links":      len(links),
		"discovered_feeds": discoveredFeeds,
	})
}

// GET /v1/crawl_jobs/{jobID}/pagerank?limit=20
func (apiCfg *apiConfig) handlerGetPageRank(w http.ResponseWriter, r *http.Request, user database.User) {
	jobID, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		respondWithError(w, 400, "Invalid job ID")
		return
	}
	job, err := apiCfg.DB.GetCrawlJobByID(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 404, "Crawl job not found")
		return
	}
	if job.UserID != user.ID {
		respondWithError(w, 403, "Forbidden")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	pages, err := apiCfg.DB.GetPagesByCrawlJob(r.Context(), jobID)
	if err != nil {
		respondWithError(w, 500, "Failed to fetch pages")
		return
	}

	type pageEntry struct {
		URL          string  `json:"url"`
		PageRank     float64 `json:"page_rank"`
		InboundCount int32   `json:"inbound_count"`
		Depth        int32   `json:"depth"`
	}
	result := make([]pageEntry, 0, limit)
	for i, p := range pages {
		if i >= limit {
			break
		}
		result = append(result, pageEntry{URL: p.Url, PageRank: p.PageRank, InboundCount: p.InboundCount, Depth: p.Depth})
	}
	respondWithJSON(w, 200, result)
}

// runCrawlJob is the background goroutine that drives the crawl and persists results.
func (apiCfg *apiConfig) runCrawlJob(jobID uuid.UUID, cfg crawler.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	apiCfg.CrawlEngines.Store(jobID, cancel)
	defer func() {
		cancel()
		apiCfg.CrawlEngines.Delete(jobID)
	}()

	if _, err := apiCfg.DB.UpdateCrawlJobStarted(ctx, jobID); err != nil {
		log.Printf("crawl job %s: failed to mark started: %v", jobID, err)
		return
	}

	engine := crawler.NewEngine(cfg)

	// Progress reporter: write pages_crawled to DB every 3s while crawling
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				apiCfg.DB.UpdateCrawlJobProgress(context.Background(), database.UpdateCrawlJobProgressParams{
					ID:           jobID,
					PagesCrawled: int32(engine.PageCount()),
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	g, err := engine.RunFull(ctx)
	cancel() // stop progress reporter
	<-progressDone

	if err != nil && err != context.Canceled {
		log.Printf("crawl job %s failed: %v", jobID, err)
		apiCfg.DB.UpdateCrawlJobFailed(context.Background(), database.UpdateCrawlJobFailedParams{
			ID:       jobID,
			ErrorMsg: sql.NullString{String: err.Error(), Valid: true},
		})
		return
	}

	bg := context.Background()

	// Persist discovered RSS/Atom feeds
	if feeds := g.DiscoveredFeeds(); len(feeds) > 0 {
		apiCfg.DB.UpdateCrawlJobDiscoveredFeeds(bg, database.UpdateCrawlJobDiscoveredFeedsParams{
			ID:              jobID,
			DiscoveredFeeds: strings.Join(feeds, "\n"),
		})
	}

	if err := persistGraph(bg, apiCfg.DB, jobID, g); err != nil {
		log.Printf("crawl job %s: failed to persist graph: %v", jobID, err)
		apiCfg.DB.UpdateCrawlJobFailed(bg, database.UpdateCrawlJobFailedParams{
			ID:       jobID,
			ErrorMsg: sql.NullString{String: err.Error(), Valid: true},
		})
		return
	}

	apiCfg.DB.UpdateCrawlJobFinished(bg, database.UpdateCrawlJobFinishedParams{
		ID:           jobID,
		PagesCrawled: int32(g.NodeCount()),
	})
	log.Printf("crawl job %s completed: %d pages, %d feeds discovered", jobID, g.NodeCount(), len(g.DiscoveredFeeds()))
}

// persistGraph writes all nodes and edges to the database.
func persistGraph(ctx context.Context, db *database.Queries, jobID uuid.UUID, g *crawler.Graph) error {
	nodes := g.Nodes()
	for _, n := range nodes {
		statusCode := sql.NullInt32{}
		if n.StatusCode != 0 {
			statusCode = sql.NullInt32{Int32: int32(n.StatusCode), Valid: true}
		}
		redirectTo := sql.NullString{}
		if n.RedirectTo != "" {
			redirectTo = sql.NullString{String: n.RedirectTo, Valid: true}
		}
		if err := db.CreatePage(ctx, database.CreatePageParams{
			ID:           uuid.New(),
			CreatedAt:    time.Now().UTC(),
			CrawlJobID:   jobID,
			Url:          n.URL,
			StatusCode:   statusCode,
			Depth:        int32(n.Depth),
			PageRank:     n.PageRank,
			InboundCount: int32(n.InboundCount),
			IsRedirect:   n.IsRedirect,
			RedirectTo:   redirectTo,
		}); err != nil {
			return err
		}
	}
	for from, targets := range g.Edges() {
		for _, to := range targets {
			if err := db.CreatePageLink(ctx, database.CreatePageLinkParams{
				ID:         uuid.New(),
				CrawlJobID: jobID,
				FromUrl:    from,
				ToUrl:      to,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
