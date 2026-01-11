-- Remove roles column from league_participants
DROP INDEX IF EXISTS idx_league_participants_roles;
ALTER TABLE league_participants DROP COLUMN IF EXISTS roles;
