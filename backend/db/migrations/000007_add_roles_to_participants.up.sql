-- Add roles column to league_participants
-- Roles: director (감독), player (선수), reserve (리저브선수), engineer (엔지니어)
-- One person can have multiple roles, stored as TEXT array

ALTER TABLE league_participants
ADD COLUMN roles TEXT[] NOT NULL DEFAULT '{}';

-- Create index for efficient role queries
CREATE INDEX idx_league_participants_roles ON league_participants USING GIN (roles);

-- Add comment for documentation
COMMENT ON COLUMN league_participants.roles IS 'Array of roles: director, player, reserve, engineer';
