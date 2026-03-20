-- +goose Up
CREATE TABLE page_links (
    id           UUID PRIMARY KEY,
    crawl_job_id UUID NOT NULL REFERENCES crawl_jobs(id) ON DELETE CASCADE,
    from_url     TEXT NOT NULL,
    to_url       TEXT NOT NULL,
    UNIQUE(crawl_job_id, from_url, to_url)
);

CREATE INDEX idx_page_links_crawl_from ON page_links(crawl_job_id, from_url);
CREATE INDEX idx_page_links_crawl_to   ON page_links(crawl_job_id, to_url);

-- +goose Down
DROP TABLE page_links;
