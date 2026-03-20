-- +goose Up
ALTER TABLE crawl_jobs ADD COLUMN discovered_feeds TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE crawl_jobs DROP COLUMN discovered_feeds;
