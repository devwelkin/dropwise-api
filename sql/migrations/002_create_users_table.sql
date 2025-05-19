-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Function and Trigger to update updated_at column for users table
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_users_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_users_updated_at_column();

-- Add user_uuid column to drops table to link to the new users table
ALTER TABLE drops ADD COLUMN user_uuid UUID NULL;

-- Add foreign key constraint from drops.user_uuid to users.id
-- ON DELETE SET NULL means if a user is deleted, the user_uuid in drops table will be set to NULL.
ALTER TABLE drops
ADD CONSTRAINT fk_drops_user_uuid
FOREIGN KEY (user_uuid) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_drops_user_uuid ON drops (user_uuid);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE drops DROP CONSTRAINT IF EXISTS fk_drops_user_uuid;
ALTER TABLE drops DROP COLUMN IF EXISTS user_uuid; -- Also removes idx_drops_user_uuid if it depends on the column

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at_column();
DROP TABLE IF EXISTS users;