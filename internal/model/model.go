package model

type Match struct {
	MatchID       int  `json:"match_id"`
	TournamentID  int  `json:"tournament_id"`
	RoundNumber   int  `json:"round_number"`
	MatchPosition int  `json:"match_position"`
	Team1ID       *int `json:"team1_id,omitempty"`
	Team2ID       *int `json:"team2_id,omitempty"`
	NextMatchID   *int `json:"next_match_id,omitempty"`
	NextMatchSlot *int `json:"next_match_slot,omitempty"`
}

type Prediction struct {
	MatchID           int  `json:"match_id"`
	PredictedWinnerID *int `json:"predicted_winner_id,omitempty"`
}

type GetBracketResponse struct {
	Matches     []Match      `json:"matches"`
	Predictions []Prediction `json:"predictions"`
	IsLocked    bool         `json:"is_locked"`
	Accuracy    *float64     `json:"accuracy,omitempty"` // Percentage (0-100)
}

type SubmitBracketRequest struct {
	Username     string       `json:"username"`
	TournamentID int          `json:"tournament_id"`
	Predictions  []Prediction `json:"predictions"`
}

type UploadMatch struct {
	RoundNumber   int    `json:"round_number"`
	MatchPosition int    `json:"match_position"`
	Team1Name     string `json:"team1_name,omitempty"` // Names instead of IDs
	Team2Name     string `json:"team2_name,omitempty"`
	NextRound     *int   `json:"next_round,omitempty"` // Used to link matches
	NextPosition  *int   `json:"next_position,omitempty"`
	NextSlot      *int   `json:"next_slot,omitempty"`
}

type TournamentUploadRequest struct {
	Title     string        `json:"title"`
	StartTime string        `json:"start_time"` // e.g., RFC3339 "2026-07-01T15:00:00Z"
	Matches   []UploadMatch `json:"matches"`
}
