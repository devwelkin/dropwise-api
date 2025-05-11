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


-- name: GetDueDropsForUser :many
-- Selects drops that are due to be sent for a specific user.
-- Drops are considered due if their status is 'new'.
-- They are ordered by priority (descending) and then by added_date (ascending).
SELECT *
FROM drops
WHERE user_id = $1
  AND status = 'new'
ORDER BY priority DESC, added_date ASC
LIMIT $2;

-- name: MarkDropAsSent :one
-- Updates a drop's status to 'sent', sets the last_sent_date, and increments the send_count.
UPDATE drops
SET
    status = 'sent',
    last_sent_date = $2, -- $2 will be the timestamp when it was sent
    send_count = send_count + 1
    -- updated_at is handled by the database trigger
WHERE id = $1 -- $1 will be the drop's ID
RETURNING *;