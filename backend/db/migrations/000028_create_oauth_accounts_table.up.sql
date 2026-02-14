-- Create oauth_accounts table for social login providers
CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_username VARCHAR(255),
    provider_avatar VARCHAR(512),
    provider_email VARCHAR(255),
    access_token VARCHAR(512),
    refresh_token VARCHAR(512),
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Each provider account can only be linked once
CREATE UNIQUE INDEX idx_oauth_accounts_provider_id ON oauth_accounts(provider, provider_id);

-- Each user can only have one account per provider
CREATE UNIQUE INDEX idx_oauth_accounts_user_provider ON oauth_accounts(user_id, provider);

-- Allow OAuth-only users (no password)
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
