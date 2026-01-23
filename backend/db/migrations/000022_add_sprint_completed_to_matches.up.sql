-- Add sprint_completed field to matches table
ALTER TABLE matches ADD COLUMN sprint_completed BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN matches.sprint_completed IS 'Whether sprint race results have been entered';
