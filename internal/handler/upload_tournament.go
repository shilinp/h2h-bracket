package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"h2h-bracket/internal/constants"
	"h2h-bracket/internal/model"
)

func (h *Handler) handleUploadTournament(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.URL.Query().Get("username")

	if username != constants.TournamentAdminUsername {
		http.Error(w, "Unauthorized: Only admin can upload tournaments", http.StatusForbidden)
		return
	}

	var req model.TournamentUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var tournamentID int
	err = tx.QueryRow(ctx, `
		INSERT INTO tournaments (title, start_time) 
		VALUES ($1, $2) RETURNING tournament_id`,
		req.Title, req.StartTime).Scan(&tournamentID)
	if err != nil {
		log.Printf("Tournament creation error: %v", err)
		http.Error(w, "Failed to create tournament", http.StatusInternalServerError)
		return
	}

	teamIDs := make(map[string]int)
	for _, m := range req.Matches {
		for _, teamName := range []string{m.Team1Name, m.Team2Name} {
			if teamName == "" {
				continue
			}
			if _, exists := teamIDs[teamName]; !exists {
				var teamID int
				err = tx.QueryRow(ctx, `
					INSERT INTO teams (team_name) VALUES ($1)
					ON CONFLICT (team_name) DO UPDATE SET team_name = EXCLUDED.team_name
					RETURNING team_id`, teamName).Scan(&teamID)
				if err != nil {
					log.Printf("Team creation error (%s): %v", teamName, err)
					http.Error(w, "Failed to provision teams", http.StatusInternalServerError)
					return
				}
				teamIDs[teamName] = teamID
			}
		}
	}

	matchIDs := make(map[string]int)
	for _, m := range req.Matches {
		var t1, t2 *int
		if m.Team1Name != "" {
			id := teamIDs[m.Team1Name]
			t1 = &id
		}
		if m.Team2Name != "" {
			id := teamIDs[m.Team2Name]
			t2 = &id
		}

		var matchID int
		err = tx.QueryRow(ctx, `
			INSERT INTO matches (tournament_id, round_number, match_position, team1_id, team2_id)
			VALUES ($1, $2, $3, $4, $5) RETURNING match_id`,
			tournamentID, m.RoundNumber, m.MatchPosition, t1, t2).Scan(&matchID)
		if err != nil {
			log.Printf("Match insertion error: %v", err)
			http.Error(w, "Failed to insert matches", http.StatusInternalServerError)
			return
		}

		key := fmt.Sprintf("%d_%d", m.RoundNumber, m.MatchPosition)
		matchIDs[key] = matchID
	}

	for _, m := range req.Matches {
		if m.NextRound != nil && m.NextPosition != nil && m.NextSlot != nil {
			currentKey := fmt.Sprintf("%d_%d", m.RoundNumber, m.MatchPosition)
			nextKey := fmt.Sprintf("%d_%d", *m.NextRound, *m.NextPosition)

			currentMatchID := matchIDs[currentKey]
			nextMatchID, exists := matchIDs[nextKey]

			if !exists {
				http.Error(w, fmt.Sprintf("Next match %s not found for match %s", nextKey, currentKey), http.StatusBadRequest)
				return
			}

			_, err = tx.Exec(ctx, `
				UPDATE matches SET next_match_id = $1, next_match_slot = $2 
				WHERE match_id = $3`,
				nextMatchID, *m.NextSlot, currentMatchID)

			if err != nil {
				log.Printf("Match linking error: %v", err)
				http.Error(w, "Failed to link bracket structure", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "Tournament uploaded successfully",
		"tournament_id": tournamentID,
	})
}
