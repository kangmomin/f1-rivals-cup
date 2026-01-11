package model

import "github.com/google/uuid"

// StandingsEntry represents a single entry in the league standings
type StandingsEntry struct {
	Rank            int       `json:"rank"`
	ParticipantID   uuid.UUID `json:"participant_id"`
	UserID          uuid.UUID `json:"user_id"`
	DriverName      string    `json:"driver_name"`
	TeamName        *string   `json:"team_name,omitempty"`
	TotalPoints     float64   `json:"total_points"`
	RacePoints      float64   `json:"race_points"`
	SprintPoints    float64   `json:"sprint_points"`
	Wins            int       `json:"wins"`
	Podiums         int       `json:"podiums"`
	FastestLaps     int       `json:"fastest_laps"`
	DNFs            int       `json:"dnfs"`
	RacesCompleted  int       `json:"races_completed"`
}

// TeamStandingsEntry represents a single team entry in the standings
type TeamStandingsEntry struct {
	Rank           int     `json:"rank"`
	TeamName       string  `json:"team_name"`
	TotalPoints    float64 `json:"total_points"`
	RacePoints     float64 `json:"race_points"`
	SprintPoints   float64 `json:"sprint_points"`
	Wins           int     `json:"wins"`
	Podiums        int     `json:"podiums"`
	FastestLaps    int     `json:"fastest_laps"`
	DNFs           int     `json:"dnfs"`
	DriverCount    int     `json:"driver_count"`
}

// LeagueStandingsResponse represents the response for league standings
type LeagueStandingsResponse struct {
	LeagueID      uuid.UUID            `json:"league_id"`
	LeagueName    string               `json:"league_name"`
	Season        int                  `json:"season"`
	TotalRaces    int                  `json:"total_races"`
	Standings     []StandingsEntry     `json:"standings"`
	TeamStandings []TeamStandingsEntry `json:"team_standings"`
}
