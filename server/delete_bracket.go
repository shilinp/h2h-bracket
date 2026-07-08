package server

import (
	"context"
	"errors"
	"io"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	"h2h-bracket/server/constants"
	"h2h-bracket/server/proto"
)

var errNotFound = errors.New("bracket not found")

// HandleDeleteBracket coordinates the deletion of a user's bracket predictions.
func (app *App) HandleDeleteBracket(w http.ResponseWriter, r *http.Request) {
	req, err := parseDeleteRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = app.deletePredictionsByUsername(r.Context(), req.GetUsername())
	if err != nil {
		if errors.Is(err, errNotFound) {
			http.Error(w, "Bracket not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete bracket", http.StatusInternalServerError)
		return
	}

	if req.GetUsername() == constants.SpecialUsername {
		err = app.unlockGlobalLock(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp := &proto.DeleteBracketResponse{Status: "Bracket deleted successfully"}
	w.Header().Set("Content-Type", "application/json")
	data, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// parseDeleteRequest extracts and validates the incoming delete command payload.
func parseDeleteRequest(r *http.Request) (*proto.DeleteBracketRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("Failed to read request body")
	}

	req := &proto.DeleteBracketRequest{}
	if err := protojson.Unmarshal(body, req); err != nil {
		return nil, errors.New("Invalid JSON payload")
	}

	if req.GetUsername() == "" {
		return nil, errors.New("username is required")
	}

	return req, nil
}

// deletePredictionsByUsername handles the direct database removal of prediction records.
func (app *App) deletePredictionsByUsername(ctx context.Context, username string) error {
	tag, err := app.DB.Exec(ctx, `
		DELETE FROM match_predictions mp
		USING users u
		WHERE mp.user_id = u.user_id AND u.username = $1`,
		username)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errNotFound
	}

	return nil
}

// unlockGlobalLock deletes the globally shared submission lock
func (app *App) unlockGlobalLock(ctx context.Context) error {
	_, err := app.DB.Exec(ctx, "DELETE FROM global_settings")
	if err != nil {
		return errors.New("Failed to clear global settings")
	}
	return nil
}
