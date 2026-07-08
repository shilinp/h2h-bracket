package server

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"

	"h2h-bracket/server/constants"
	"h2h-bracket/server/proto"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

// HandleSubmitBracket controls prediction submissions and state changes inside transactions.
func (app *App) HandleSubmitBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := parseSubmitRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	username, isSpecialUser, isMaster := app.evaluateUserContext(req)
	if !isMaster && username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	if !isMaster {
		locked, err := app.checkIsLocked(ctx)
		if err != nil {
			http.Error(w, "Failed to check lock status", http.StatusInternalServerError)
			return
		}
		if locked {
			http.Error(w, "Submissions are locked. The special bracket has been finalized.", http.StatusForbidden)
			return
		}
	}

	if err := app.processBracketSubmission(ctx, username, isMaster, req.GetPredictions()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedBracket, err := app.fetchActiveTournamentBracket(ctx, username, isSpecialUser)
	if err != nil {
		http.Error(w, "Failed to fetch updated bracket", http.StatusInternalServerError)
		return
	}

	resp := &proto.SubmitBracketResponse{
		Status:         "Predictions saved successfully",
		UpdatedBracket: updatedBracket,
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// parseSubmitRequest unmarshals incoming request text strings into proper payload definitions.
func parseSubmitRequest(r *http.Request) (*proto.SubmitBracketRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("Failed to read request body")
	}

	req := &proto.SubmitBracketRequest{}
	if err := protojson.Unmarshal(body, req); err != nil {
		return nil, errors.New("Invalid JSON payload")
	}
	return req, nil
}

// evaluateUserContext returns routing details based on administrative permission configurations.
func (app *App) evaluateUserContext(req *proto.SubmitBracketRequest) (string, bool, bool) {
	username := req.GetUsername()
	isSpecialUser := req.GetIsSpecialUser()
	if isSpecialUser {
		username = constants.SpecialUsername
	}
	isMaster := isSpecialUser || username == constants.SpecialUsername
	return username, isSpecialUser, isMaster
}

// checkIsLocked queries state definitions to determine submission window availability.
func (app *App) checkIsLocked(ctx context.Context) (bool, error) {
	var isLocked bool
	err := app.DB.QueryRow(ctx, "SELECT is_locked FROM global_settings LIMIT 1").Scan(&isLocked)
	if err != nil && err != pgx.ErrNoRows {
		return false, err
	}
	return isLocked, nil
}

// processBracketSubmission writes and flushes standard transactions into datastores.
func (app *App) processBracketSubmission(ctx context.Context, username string, isMaster bool, predictions map[int32]int32) error {
	tx, err := app.DB.Begin(ctx)
	if err != nil {
		return errors.New("Failed to start transaction")
	}
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username) VALUES ($1) 
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username 
		RETURNING user_id`, username).Scan(&userID)
	if err != nil {
		return errors.New("Failed to resolve user")
	}

	_, err = tx.Exec(ctx, "DELETE FROM match_predictions WHERE user_id = $1", userID)
	if err != nil {
		return errors.New("Failed to clear old predictions")
	}

	batch := &pgx.Batch{}
	for matchID, winnerID := range predictions {
		batch.Queue(`
			INSERT INTO match_predictions (user_id, match_id, predicted_winner_id) 
			VALUES ($1, $2, $3)`,
			userID, int(matchID), int(winnerID))
	}

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		log.Printf("batch insert failed: %v", err)
		return errors.New("Failed to save predictions")
	}

	if isMaster {
		_, _ = tx.Exec(ctx, "DELETE FROM global_settings")
		_, _ = tx.Exec(ctx, "INSERT INTO global_settings (is_locked) VALUES (TRUE)")
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.New("Failed to commit transaction")
	}

	return nil
}
