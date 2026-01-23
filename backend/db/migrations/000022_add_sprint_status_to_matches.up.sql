-- Add sprint_status column with default value
ALTER TABLE matches ADD COLUMN IF NOT EXISTS sprint_status VARCHAR(20) DEFAULT 'upcoming';

-- Migrate existing data: convert sprint_completed boolean to sprint_status
UPDATE matches SET sprint_status = CASE
    WHEN sprint_completed = true THEN 'completed'
    ELSE 'upcoming'
END WHERE sprint_completed IS NOT NULL;

-- Drop the old column
ALTER TABLE matches DROP COLUMN IF EXISTS sprint_completed;

COMMENT ON COLUMN matches.sprint_status IS 'Sprint race status (upcoming, in_progress, completed, cancelled)';
