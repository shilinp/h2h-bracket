package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"h2h-bracket/internal/constants"
	"h2h-bracket/internal/model"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) handleSubmitBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.SubmitBracketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Username != constants.PlayerAdminUsername {
		var isLocked bool
		h.DB.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM user_brackets ub
				JOIN users u ON u.user_id = ub.user_id
				WHERE u.username = $1 AND ub.tournament_id = $2
			)`, constants.PlayerAdminUsername, req.TournamentID).Scan(&isLocked)

		if isLocked {
			http.Error(w, "Submissions are locked. The admin has finalized the root bracket.", http.StatusForbidden)
			return
		}
	}

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username) VALUES ($1) 
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username 
		RETURNING user_id`, req.Username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to resolve user", http.StatusInternalServerError)
		return
	}

	var userBracketID int
	err = tx.QueryRow(ctx, `
		INSERT INTO user_brackets (user_id, tournament_id) VALUES ($1, $2)
		ON CONFLICT (user_id, tournament_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING user_bracket_id`, userID, req.TournamentID).Scan(&userBracketID)
	if err != nil {
		log.Printf("Error inserting into user_brackets: %v\n", err)
		http.Error(w, "Failed to resolve bracket", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx, "DELETE FROM match_predictions WHERE user_bracket_id = $1", userBracketID)
	if err != nil {
		http.Error(w, "Failed to clear old predictions", http.StatusInternalServerError)
		return
	}

	batch := &pgx.Batch{}
	for _, p := range req.Predictions {
		batch.Queue(`
			INSERT INTO match_predictions (user_bracket_id, match_id, predicted_winner_id) 
			VALUES ($1, $2, $3)`,
			userBracketID, p.MatchID, p.PredictedWinnerID)
	}

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		log.Printf("batch insert failed: %v", err)
		http.Error(w, "Failed to save predictions", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Predictions saved successfully"})
}
