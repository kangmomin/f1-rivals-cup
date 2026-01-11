-- Create matches table for league schedules
CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    round INT NOT NULL,
    track VARCHAR(100) NOT NULL,
    match_date DATE NOT NULL,
    match_time TIME,
    status VARCHAR(20) NOT NULL DEFAULT 'upcoming',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(league_id, round)
);

-- Create indexes
CREATE INDEX idx_matches_league_id ON matches(league_id);
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_matches_match_date ON matches(match_date);

-- Add comments
COMMENT ON COLUMN matches.status IS 'Match status: upcoming, in_progress, completed, cancelled';
