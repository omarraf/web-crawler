-- +goose Up
ALTER TABLE feeds ADD COLUMN last_fetched_at TIMESTAMP;

-- +goose Down
-- Temporarily commented to fix migration issue
-- ALTER TABLE feeds DROP COLUMN last_fetched_at;
