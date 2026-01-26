-- Team change requests table
CREATE TABLE IF NOT EXISTS team_change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id UUID NOT NULL REFERENCES league_participants(id) ON DELETE CASCADE,
    current_team_name VARCHAR(100),
    requested_team_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reason TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Participant team history table for tracking team changes over time
CREATE TABLE IF NOT EXISTS participant_team_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_id UUID NOT NULL REFERENCES league_participants(id) ON DELETE CASCADE,
    team_name VARCHAR(100) NOT NULL,
    effective_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMP WITH TIME ZONE,
    change_request_id UUID REFERENCES team_change_requests(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_team_change_requests_participant_id ON team_change_requests(participant_id);
CREATE INDEX idx_team_change_requests_status ON team_change_requests(status);
CREATE INDEX idx_team_change_requests_requested_team ON team_change_requests(requested_team_name);
CREATE INDEX idx_participant_team_history_participant_id ON participant_team_history(participant_id);
CREATE INDEX idx_participant_team_history_effective_dates ON participant_team_history(effective_from, effective_until);

-- Migrate existing data: create initial team history records for participants with teams
INSERT INTO participant_team_history (participant_id, team_name, effective_from)
SELECT id, team_name, created_at
FROM league_participants
WHERE team_name IS NOT NULL AND team_name != '';

-- Add constraint to prevent duplicate pending requests
CREATE UNIQUE INDEX idx_team_change_requests_pending_unique
ON team_change_requests(participant_id)
WHERE status = 'pending';
