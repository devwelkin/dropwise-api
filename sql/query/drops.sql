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