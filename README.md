# Web Crawler

A web crawler and RSS aggregator built with Go. Crawl any site from a seed URL, visualize its link graph, and subscribe to discovered RSS feeds — all from a single interface.

## Features

- **BFS web crawler** — concurrent worker pool, configurable depth and page limits, rate limiting, live progress tracking
- **Graph analysis** — PageRank scoring, orphaned pages, broken links, redirect chains, link authority rankings
- **RSS aggregator** — follow feeds, background scraping, post timeline
- **Feed discovery** — automatically detects RSS/Atom feeds during crawl; subscribe in one click
- **D3.js visualization** — force-directed graph with nodes sized by PageRank, colored by depth, zoom/pan, click-to-highlight
- **REST API** — API key auth, 15+ endpoints, all responses typed via sqlc-generated queries

## Stack

- **Go** — crawler engine, REST API (Chi router)
- **PostgreSQL** — persistence (Goose migrations, sqlc)
- **D3.js** — graph visualization
- **Neon** — serverless Postgres hosting

## Setup

**Prerequisites:** Go 1.21+, PostgreSQL, [Goose](https://github.com/pressly/goose), [sqlc](https://sqlc.dev)

```bash
git clone https://github.com/omarraf/web-scraper.git
cd web-scraper
go mod vendor
```

Create a `.env` file:
```
PORT=8000
DATABASE_URL=your_postgres_connection_string
```

Run migrations and start:
```bash
goose -dir sql/schema postgres $DATABASE_URL up
go run .
```

Open `http://localhost:8000` for the UI.

## API

All routes under `/v1`. Authenticated routes require `Authorization: ApiKey <key>`.

```
POST   /v1/users                          Create user (returns API key)
GET    /v1/users                          Get current user

POST   /v1/feeds                          Add RSS feed
GET    /v1/feeds                          List all feeds
POST   /v1/feed_follows                   Follow a feed
DELETE /v1/feed_follows/{id}              Unfollow a feed
GET    /v1/posts                          Get posts from followed feeds

POST   /v1/crawl_jobs                     Start a crawl
GET    /v1/crawl_jobs                     List crawl jobs
GET    /v1/crawl_jobs/{id}                Job status + progress
DELETE /v1/crawl_jobs/{id}               Cancel running crawl
GET    /v1/crawl_jobs/{id}/graph          Nodes + edges for visualization
GET    /v1/crawl_jobs/{id}/analysis       Orphans, broken links, redirects, discovered feeds
GET    /v1/crawl_jobs/{id}/pagerank       Top pages by PageRank
```

## Quick start

```bash
# Create a user
curl -X POST http://localhost:8000/v1/users -d '{"name":"alice"}'
# → copy the api_key from the response

# Start a crawl
curl -X POST http://localhost:8000/v1/crawl_jobs \
  -H "Authorization: ApiKey <key>" \
  -d '{"seed_url":"https://example.com","max_depth":2,"max_pages":100}'

# Or just open http://localhost:8000 and use the UI
```
