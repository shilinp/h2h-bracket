package handler

import (
	"io"
	"log"
	"net/http"

	"google.golang.org/protobuf/proto"

	"h2h-bracket/internal/constants"
	pb "h2h-bracket/internal/proto"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) handleSubmitBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	req := &pb.SubmitBracketRequest{}
	if err := proto.Unmarshal(body, req); err != nil {
		http.Error(w, "Invalid protobuf payload", http.StatusBadRequest)
		return
	}

	username := req.GetUsername()
	isSpecialUser := req.GetIsSpecialUser()
	if isSpecialUser {
		username = constants.SpecialUsername
	}

	if !isSpecialUser && username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	isMaster := isSpecialUser || username == constants.SpecialUsername
	if !isMaster {
		var isLocked bool
		err = h.DB.QueryRow(ctx, "SELECT is_locked FROM global_settings LIMIT 1").Scan(&isLocked)
		if err != nil && err != pgx.ErrNoRows {
			http.Error(w, "Failed to check lock status", http.StatusInternalServerError)
			return
		}

		if isLocked {
			http.Error(w, "Submissions are locked. The special bracket has been finalized.", http.StatusForbidden)
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
		RETURNING user_id`, username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to resolve user", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(ctx, "DELETE FROM match_predictions WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to clear old predictions", http.StatusInternalServerError)
		return
	}

	batch := &pgx.Batch{}
	for matchID, winnerID := range req.GetPredictions() {
		batch.Queue(`
			INSERT INTO match_predictions (user_id, match_id, predicted_winner_id) 
			VALUES ($1, $2, $3)`,
			userID, int(matchID), int(winnerID))
	}

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		log.Printf("batch insert failed: %v", err)
		http.Error(w, "Failed to save predictions", http.StatusInternalServerError)
		return
	}

	if isMaster {
		_, _ = tx.Exec(ctx, "DELETE FROM global_settings")
		_, _ = tx.Exec(ctx, "INSERT INTO global_settings (is_locked) VALUES (TRUE)")
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	updatedBracket, err := h.fetchActiveTournamentBracket(ctx, username, isSpecialUser)
	if err != nil {
		http.Error(w, "Failed to fetch updated bracket", http.StatusInternalServerError)
		return
	}

	resp := &pb.SubmitBracketResponse{
		Status:         "Predictions saved successfully",
		UpdatedBracket: updatedBracket,
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
