-- name: AddTagToDrop :exec
-- Associates a tag with a drop.
-- ON CONFLICT DO NOTHING prevents errors if the association already exists.
INSERT INTO drops_item_tags (drops_id, tag_id)
VALUES ($1, $2)
ON CONFLICT (drops_id, tag_id) DO NOTHING;

-- name: GetTagsForDrop :many
-- Retrieves all tags associated with a specific drop.
SELECT t.id, t.name
FROM tags t
JOIN drops_item_tags dit ON t.id = dit.tag_id
WHERE dit.drops_id = $1
ORDER BY t.name;

-- name: RemoveTagFromDrop :exec
-- Removes a specific tag association from a drop.
DELETE FROM drops_item_tags
WHERE drops_id = $1 AND tag_id = $2;

-- name: RemoveAllTagsFromDrop :exec
-- Removes all tag associations for a specific drop.
-- Useful when updating a drop's tags to clear existing ones first.
DELETE FROM drops_item_tags
WHERE drops_id = $1;