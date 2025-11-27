-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeeds :many
SELECT
    f.url AS feed_url,
    f.name AS feed_name,
    u.name AS user_name
FROM feeds f INNER JOIN users u
ON f.user_id = u.id;
