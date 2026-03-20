-- +goose Up
CREATE TABLE crawl_jobs (
    id            UUID PRIMARY KEY,
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seed_url      TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    max_depth     INT NOT NULL DEFAULT 3,
    max_pages     INT NOT NULL DEFAULT 500,
    pages_crawled INT NOT NULL DEFAULT 0,
    started_at    TIMESTAMP,
    finished_at   TIMESTAMP,
    error_msg     TEXT
);

-- +goose Down
DROP TABLE crawl_jobs;
