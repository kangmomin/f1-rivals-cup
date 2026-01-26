-- Add requested_roles column to team_change_requests
-- Stores the roles the participant wants to have in the new team
ALTER TABLE team_change_requests
ADD COLUMN requested_roles TEXT[] DEFAULT NULL;

-- Add current_roles column to store the roles at the time of request
ALTER TABLE team_change_requests
ADD COLUMN current_roles TEXT[] DEFAULT NULL;
