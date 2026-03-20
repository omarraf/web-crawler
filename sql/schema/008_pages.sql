-- +goose Up
CREATE TABLE pages (
    id            UUID PRIMARY KEY,
    created_at    TIMESTAMP NOT NULL,
    crawl_job_id  UUID NOT NULL REFERENCES crawl_jobs(id) ON DELETE CASCADE,
    url           TEXT NOT NULL,
    status_code   INT,
    depth         INT NOT NULL DEFAULT 0,
    page_rank     DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    inbound_count INT NOT NULL DEFAULT 0,
    is_redirect   BOOLEAN NOT NULL DEFAULT FALSE,
    redirect_to   TEXT,
    UNIQUE(crawl_job_id, url)
);

-- +goose Down
DROP TABLE pages;
