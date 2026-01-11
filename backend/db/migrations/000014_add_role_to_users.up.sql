-- Add role and permissions columns to users table
-- Role: USER (default), STAFF, ADMIN
-- Permissions: JSONB array with GIN index for fast lookup

-- Add role column with enum-like constraint
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'USER';

-- Add permissions JSONB column
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Add version column for optimistic locking
ALTER TABLE users ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_permissions ON users USING GIN (permissions);

-- Add check constraint for valid roles
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('USER', 'STAFF', 'ADMIN'));
