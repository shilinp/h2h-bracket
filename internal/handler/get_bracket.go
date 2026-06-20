package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"h2h-bracket/internal/constants"
	"h2h-bracket/internal/model"
)

func (h *Handler) handleGetBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.URL.Query().Get("username")
	tournamentIDStr := r.URL.Query().Get("tournament_id")

	if username == "" || tournamentIDStr == "" {
		http.Error(w, "username and tournament_id are required", http.StatusBadRequest)
		return
	}

	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		http.Error(w, "invalid tournament_id", http.StatusBadRequest)
		return
	}

	resp := model.GetBracketResponse{
		Matches:     []model.Match{},
		Predictions: []model.Prediction{},
	}

	rows, err := h.DB.Query(ctx, `
		SELECT match_id, tournament_id, round_number, match_position, team1_id, team2_id, next_match_id, next_match_slot 
		FROM matches WHERE tournament_id = $1 ORDER BY round_number, match_position`, tournamentID)
	if err != nil {
		http.Error(w, "Database error fetching matches", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m model.Match
		err := rows.Scan(&m.MatchID, &m.TournamentID, &m.RoundNumber, &m.MatchPosition, &m.Team1ID, &m.Team2ID, &m.NextMatchID, &m.NextMatchSlot)
		if err != nil {
			log.Println("Error scanning match:", err)
			continue
		}
		resp.Matches = append(resp.Matches, m)
	}

	h.DB.Exec(ctx, `INSERT INTO users (username) VALUES ($1) ON CONFLICT (username) DO NOTHING`, username)

	var userBracketID int
	err = h.DB.QueryRow(ctx, `
		SELECT ub.user_bracket_id FROM user_brackets ub
		JOIN users u ON ub.user_id = u.user_id
		WHERE u.username = $1 AND ub.tournament_id = $2`, username, tournamentID).Scan(&userBracketID)

	if err == nil {
		pRows, _ := h.DB.Query(ctx, `
			SELECT match_id, predicted_winner_id 
			FROM match_predictions WHERE user_bracket_id = $1`, userBracketID)
		defer pRows.Close()

		for pRows.Next() {
			var p model.Prediction
			pRows.Scan(&p.MatchID, &p.PredictedWinnerID)
			resp.Predictions = append(resp.Predictions, p)
		}
	}

	var adminBracketID int
	err = h.DB.QueryRow(ctx, `
		SELECT ub.user_bracket_id FROM user_brackets ub
		JOIN users u ON ub.user_id = u.user_id
		WHERE u.username = $1 AND ub.tournament_id = $2`, constants.PlayerAdminUsername, tournamentID).Scan(&adminBracketID)

	if err == nil {
		resp.IsLocked = true
		if username != constants.PlayerAdminUsername && userBracketID != 0 {
			var total, correct int
			err = h.DB.QueryRow(ctx, `
				SELECT 
					COUNT(*) as total,
					SUM(CASE WHEN up.predicted_winner_id = ap.predicted_winner_id THEN 1 ELSE 0 END) as correct
				FROM match_predictions up
				JOIN match_predictions ap ON up.match_id = ap.match_id
				WHERE up.user_bracket_id = $1 AND ap.user_bracket_id = $2`,
				userBracketID, adminBracketID).Scan(&total, &correct)

			if err == nil && total > 0 {
				acc := (float64(correct) / float64(total)) * 100.0
				resp.Accuracy = &acc
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
