-- name: CreateDrop :one
INSERT INTO drops (
    user_id,
    topic,
    url,
    user_notes,
    priority
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;


-- name: GetDrop :one
SELECT * FROM drops
WHERE id = $1;


-- name: ListDrops :many
SELECT * FROM drops
WHERE user_id = $1
ORDER BY added_date DESC;


-- name: UpdateDrop :one
UPDATE drops
SET
    topic = COALESCE(sqlc.narg('topic'), topic),
    url = COALESCE(sqlc.narg('url'), url),
    user_notes = COALESCE(sqlc.narg('user_notes'), user_notes),
    priority = COALESCE(sqlc.narg('priority'), priority),
    status = COALESCE(sqlc.narg('status'), status)
    -- updated_at is handled by the database trigger
WHERE id = $1 AND user_id = $2
RETURNING *;


-- name: DeleteDrop :exec
DELETE FROM drops
WHERE id = $1 AND user_id = $2;