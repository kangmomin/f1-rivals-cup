-- Create match_results table for storing race results
CREATE TABLE IF NOT EXISTS match_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES league_participants(id) ON DELETE CASCADE,
    position INT,
    points DECIMAL(5,1) NOT NULL DEFAULT 0,
    fastest_lap BOOLEAN NOT NULL DEFAULT false,
    dnf BOOLEAN NOT NULL DEFAULT false,
    dnf_reason VARCHAR(100),
    sprint_position INT,
    sprint_points DECIMAL(5,1) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(match_id, participant_id)
);

-- Create indexes
CREATE INDEX idx_match_results_match_id ON match_results(match_id);
CREATE INDEX idx_match_results_participant_id ON match_results(participant_id);
CREATE INDEX idx_match_results_position ON match_results(position);

-- Add comments
COMMENT ON TABLE match_results IS 'Stores race results for each match';
COMMENT ON COLUMN match_results.position IS 'Final race position (null if DNF)';
COMMENT ON COLUMN match_results.points IS 'Points earned from main race';
COMMENT ON COLUMN match_results.fastest_lap IS 'Whether driver set the fastest lap (bonus point)';
COMMENT ON COLUMN match_results.dnf IS 'Did Not Finish flag';
COMMENT ON COLUMN match_results.dnf_reason IS 'Reason for DNF (crash, mechanical, etc.)';
COMMENT ON COLUMN match_results.sprint_position IS 'Sprint race position (if applicable)';
COMMENT ON COLUMN match_results.sprint_points IS 'Points earned from sprint race';
