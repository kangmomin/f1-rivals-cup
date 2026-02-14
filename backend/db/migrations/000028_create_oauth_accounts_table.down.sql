-- Set password_hash to empty string for OAuth-only users before restoring NOT NULL
UPDATE users SET password_hash = '' WHERE password_hash IS NULL;

-- Restore NOT NULL constraint on password_hash
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

-- Drop oauth_accounts table
DROP TABLE IF EXISTS oauth_accounts;
