-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- TODO: Add a query to get a user by their API key

-- name: GetUserByAPIKey :one
SELECT * FROM users WHERE api_key = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;




