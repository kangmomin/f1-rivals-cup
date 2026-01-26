-- Team change activity log table for audit purposes
CREATE TABLE IF NOT EXISTS team_change_activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NOT NULL REFERENCES users(id),
    request_id UUID NOT NULL REFERENCES team_change_requests(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES league_participants(id) ON DELETE CASCADE,
    action_type VARCHAR(20) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_tcal_request ON team_change_activity_log(request_id, created_at DESC);
CREATE INDEX idx_tcal_participant ON team_change_activity_log(participant_id, created_at DESC);

-- Constraint for valid action types
ALTER TABLE team_change_activity_log ADD CONSTRAINT chk_tcal_action_type
    CHECK (action_type IN ('CREATE', 'APPROVE', 'REJECT', 'CANCEL'));
