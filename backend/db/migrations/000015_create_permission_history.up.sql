-- Permission History table for audit trail
-- Tracks all role and permission changes

CREATE TABLE IF NOT EXISTS permission_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Who changed
    changer_id UUID NOT NULL REFERENCES users(id),

    -- Who was changed
    target_id UUID NOT NULL REFERENCES users(id),

    -- What changed
    change_type VARCHAR(20) NOT NULL, -- 'ROLE' or 'PERMISSION'
    old_value JSONB NOT NULL,
    new_value JSONB NOT NULL,

    -- When
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for querying history by target user (most common query)
CREATE INDEX idx_permission_history_target ON permission_history(target_id, created_at DESC);

-- Index for querying by changer
CREATE INDEX idx_permission_history_changer ON permission_history(changer_id, created_at DESC);

-- Add check constraint for change_type
ALTER TABLE permission_history ADD CONSTRAINT chk_permission_history_type
    CHECK (change_type IN ('ROLE', 'PERMISSION'));

-- Comment for documentation
COMMENT ON TABLE permission_history IS 'Audit log for role and permission changes. Retained for 3 years.';
