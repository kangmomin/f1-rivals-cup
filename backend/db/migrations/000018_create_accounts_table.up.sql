CREATE TABLE accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
  owner_id UUID NOT NULL,
  owner_type VARCHAR(20) NOT NULL, -- 'team', 'participant', 'system'
  balance BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(league_id, owner_id, owner_type)
);

CREATE INDEX idx_accounts_league_id ON accounts(league_id);
CREATE INDEX idx_accounts_owner ON accounts(owner_id, owner_type);
