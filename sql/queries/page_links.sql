-- name: CreatePageLink :exec
INSERT INTO page_links (id, crawl_job_id, from_url, to_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (crawl_job_id, from_url, to_url) DO NOTHING;

-- name: GetPageLinksByCrawlJob :many
SELECT * FROM page_links WHERE crawl_job_id = $1;
