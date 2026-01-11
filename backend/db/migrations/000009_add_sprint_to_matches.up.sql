-- Add sprint race fields to matches table
ALTER TABLE matches ADD COLUMN has_sprint BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE matches ADD COLUMN sprint_time TIME;

COMMENT ON COLUMN matches.has_sprint IS 'Whether this round includes a sprint race';
COMMENT ON COLUMN matches.sprint_time IS 'Sprint race time (if has_sprint is true)';
