-- name: CreateTag :one
-- Upsert a tag: inserts a new tag if the name doesn't exist,
-- or returns the existing tag if the name matches.
-- The DO UPDATE clause is necessary to make RETURNING * work consistently for both insert and conflict cases.
INSERT INTO tags (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: GetTagByName :one
SELECT * FROM tags
WHERE name = $1;

-- name: ListTags :many
SELECT * FROM tags
ORDER BY name;