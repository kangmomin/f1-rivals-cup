-- Create refresh_tokens table for multi-device support
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(512) NOT NULL UNIQUE,
    device_info VARCHAR(255),  -- Optional: browser/device identifier
    ip_address VARCHAR(45),    -- IPv4 or IPv6
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Migrate existing refresh tokens from users table to refresh_tokens table
INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, last_used_at)
SELECT
    id as user_id,
    refresh_token as token,
    NOW() + INTERVAL '7 days' as expires_at,  -- Default expiry
    updated_at as created_at,
    updated_at as last_used_at
FROM users
WHERE refresh_token IS NOT NULL AND refresh_token != '';
