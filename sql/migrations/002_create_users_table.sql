
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

-- Add foreign key constraint from drops.user_uuid to users.id
-- The user_uuid column itself is now created in 001_create_initial_schema.sql
-- ON DELETE SET NULL means if a user is deleted, the user_uuid in drops table will be set to NULL.
ALTER TABLE drops
ADD CONSTRAINT fk_drops_user_uuid
FOREIGN KEY (user_uuid) REFERENCES users(id) ON DELETE SET NULL;

-- The index idx_drops_user_uuid is removed as idx_drops_user_uuid_status from 001 covers its primary use.

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE drops DROP CONSTRAINT IF EXISTS fk_drops_user_uuid;
-- Note: We don't drop the user_uuid column here in the down migration of 002,
-- because it's defined in 001. The down migration of 001 will drop the drops table.

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at_column();
DROP TABLE IF EXISTS users;