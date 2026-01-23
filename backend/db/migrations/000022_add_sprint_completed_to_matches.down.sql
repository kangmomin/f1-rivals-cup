-- Remove sprint_completed field from matches table
ALTER TABLE matches DROP COLUMN IF EXISTS sprint_completed;
