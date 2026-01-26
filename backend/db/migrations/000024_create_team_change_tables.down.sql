-- Drop indexes
DROP INDEX IF EXISTS idx_team_change_requests_pending_unique;
DROP INDEX IF EXISTS idx_participant_team_history_effective_dates;
DROP INDEX IF EXISTS idx_participant_team_history_participant_id;
DROP INDEX IF EXISTS idx_team_change_requests_requested_team;
DROP INDEX IF EXISTS idx_team_change_requests_status;
DROP INDEX IF EXISTS idx_team_change_requests_participant_id;

-- Drop tables
DROP TABLE IF EXISTS participant_team_history;
DROP TABLE IF EXISTS team_change_requests;
