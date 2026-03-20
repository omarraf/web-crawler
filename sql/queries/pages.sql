-- name: CreatePage :exec
INSERT INTO pages (id, created_at, crawl_job_id, url, status_code, depth, page_rank, inbound_count, is_redirect, redirect_to)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (crawl_job_id, url) DO NOTHING;

-- name: GetPagesByCrawlJob :many
SELECT * FROM pages WHERE crawl_job_id = $1 ORDER BY page_rank DESC;
