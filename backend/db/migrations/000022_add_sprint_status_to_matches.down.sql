-- Revert sprint_status to sprint_completed
ALTER TABLE matches DROP COLUMN IF EXISTS sprint_status;
ALTER TABLE matches ADD COLUMN sprint_completed BOOLEAN NOT NULL DEFAULT false;
