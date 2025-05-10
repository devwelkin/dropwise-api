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