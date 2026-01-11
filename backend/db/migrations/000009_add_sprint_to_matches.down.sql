-- Remove sprint race fields from matches table
ALTER TABLE matches DROP COLUMN IF EXISTS sprint_time;
ALTER TABLE matches DROP COLUMN IF EXISTS has_sprint;
