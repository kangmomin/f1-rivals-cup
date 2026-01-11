CREATE TABLE IF NOT EXISTS league_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    team_name VARCHAR(100),
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(league_id, user_id)
);

CREATE INDEX idx_league_participants_league_id ON league_participants(league_id);
CREATE INDEX idx_league_participants_user_id ON league_participants(user_id);
CREATE INDEX idx_league_participants_status ON league_participants(status);
