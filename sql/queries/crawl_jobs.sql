-- name: CreateCrawlJob :one
INSERT INTO crawl_jobs (id, created_at, updated_at, user_id, seed_url, status, max_depth, max_pages)
VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7)
RETURNING *;

-- name: GetCrawlJobByID :one
SELECT * FROM crawl_jobs WHERE id = $1;

-- name: GetCrawlJobsByUserID :many
SELECT * FROM crawl_jobs WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateCrawlJobStarted :one
UPDATE crawl_jobs
SET status = 'running', started_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCrawlJobFinished :one
UPDATE crawl_jobs
SET status = 'completed', finished_at = NOW(), updated_at = NOW(), pages_crawled = $2
WHERE id = $1
RETURNING *;

-- name: UpdateCrawlJobFailed :one
UPDATE crawl_jobs
SET status = 'failed', finished_at = NOW(), updated_at = NOW(), error_msg = $2
WHERE id = $1
RETURNING *;

-- name: UpdateCrawlJobDiscoveredFeeds :exec
UPDATE crawl_jobs SET discovered_feeds = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateCrawlJobProgress :exec
UPDATE crawl_jobs SET pages_crawled = $2, updated_at = NOW() WHERE id = $1;
