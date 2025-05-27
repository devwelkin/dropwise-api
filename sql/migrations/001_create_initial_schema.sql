
-- +goose Up
CREATE TABLE drops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uuid UUID NULL, -- Changed from user_id VARCHAR(255), allowing NULL for now
    topic TEXT NOT NULL,
    url TEXT NOT NULL,
    user_notes TEXT,
    added_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'sent', 'archived', 'snoozed')),
    last_sent_date TIMESTAMPTZ,
    send_count INTEGER NOT NULL DEFAULT 0,
    priority INTEGER DEFAULT 0
);

-- Changed index to use user_uuid
CREATE INDEX idx_drops_user_uuid_status ON drops (user_uuid, status);
CREATE INDEX idx_drops_status_last_sent ON drops (status, last_sent_date);

CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE drops_item_tags (
    drops_id UUID NOT NULL REFERENCES drops(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (drops_id, tag_id)
);

CREATE INDEX idx_drops_item_tags_tag_id ON drops_item_tags (tag_id);

-- Function and Trigger to update updated_at column
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

CREATE TRIGGER update_drops_updated_at
BEFORE UPDATE ON drops
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TRIGGER IF EXISTS update_drops_updated_at ON drops;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS drops_item_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS drops; -- This will also drop idx_drops_user_uuid_status and idx_drops_status_last_sent