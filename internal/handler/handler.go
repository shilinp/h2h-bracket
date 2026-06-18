package handler

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"h2h-bracket/internal/constants"
	"h2h-bracket/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	assets embed.FS
	DB     *pgxpool.Pool
}

func New(assets embed.FS, db *pgxpool.Pool) *Handler {
	return &Handler{assets: assets, DB: db}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	publicFS, _ := fs.Sub(h.assets, "public")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(publicFS))))

	mux.HandleFunc("GET /api/bracket", h.handleGetBracket)
	mux.HandleFunc("POST /api/bracket", h.handleSubmitBracket)
	mux.HandleFunc("DELETE /api/admin/bracket", h.handleDeleteAdminBracket)
	mux.HandleFunc("POST /api/admin/tournament/upload", h.handleUploadTournament)
}

// --- Handlers ---

// GET /api/bracket?username={user}&tournament_id={id}
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

	// 1. Fetch initial matchups
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

	// 2. Fetch user's existing predictions (if any)
	h.ensureUserExists(ctx, username)

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

	// 3. Check for Admin Root Bracket & Calculate Accuracy
	var adminBracketID int
	err = h.DB.QueryRow(ctx, `
		SELECT ub.user_bracket_id FROM user_brackets ub
		JOIN users u ON ub.user_id = u.user_id
		WHERE u.username = $1 AND ub.tournament_id = $2`, constants.PlayerAdminUsername, tournamentID).Scan(&adminBracketID)

	if err == nil {
		resp.IsLocked = true // Admin bracket exists, submissions are locked

		// Only calculate accuracy if the user is NOT the admin and has a bracket
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

// POST /api/bracket
func (h *Handler) handleSubmitBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.SubmitBracketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Check if admin bracket exists to enforce lock
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

	// Start a transaction for safe upsert
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Auto-vivify user
	var userID int
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username) VALUES ($1) 
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username 
		RETURNING user_id`, req.Username).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to resolve user", http.StatusInternalServerError)
		return
	}

	// Upsert User Bracket
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

	// Clear existing predictions for a clean overwrite, or bulk upsert.
	// Deleting and re-inserting is generally safer for "resubmissions" of arrays.
	_, err = tx.Exec(ctx, "DELETE FROM match_predictions WHERE user_bracket_id = $1", userBracketID)
	if err != nil {
		http.Error(w, "Failed to clear old predictions", http.StatusInternalServerError)
		return
	}

	// Insert new predictions
	batch := &pgx.Batch{}
	for _, p := range req.Predictions {
		batch.Queue(`
			INSERT INTO match_predictions (user_bracket_id, match_id, predicted_winner_id) 
			VALUES ($1, $2, $3, $4, $5)`,
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

// DELETE /api/admin/bracket?tournament_id={id}
func (h *Handler) handleDeleteAdminBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tournamentIDStr := r.URL.Query().Get("tournament_id")

	// CASCADE delete handles the match_predictions automatically based on the schema setup
	tag, err := h.DB.Exec(ctx, `
		DELETE FROM user_brackets ub
		USING users u
		WHERE ub.user_id = u.user_id AND u.username = $1 AND ub.tournament_id = $2`,
		constants.PlayerAdminUsername, tournamentIDStr)

	if err != nil {
		http.Error(w, "Failed to delete root bracket", http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() == 0 {
		http.Error(w, "Root bracket not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "Root bracket deleted, submissions unlocked"})
}

// POST /api/admin/tournament/upload?username=admin
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

	// 1. Create the Tournament
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

	// 2. Resolve Teams (Insert if missing, get IDs)
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

	// 3. First Pass: Insert Matches (without next_match links)
	// We map "Round_Position" -> Match_ID so we can link them later
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

	// 4. Second Pass: Link Matches (next_match_id)
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

// --- Helpers ---

func (h *Handler) ensureUserExists(ctx context.Context, username string) {
	h.DB.Exec(ctx, `INSERT INTO users (username) VALUES ($1) ON CONFLICT (username) DO NOTHING`, username)
}
