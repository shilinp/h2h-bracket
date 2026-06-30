package handler

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "h2h-bracket/internal/proto"
)

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p *= 2
	}
	return p
}

func shuffle(slice []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(slice) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func (h *Handler) handleUploadTournament(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	req := &pb.TournamentUploadRequest{}
	
	// FIX 1: Support JSON payloads from the Svelte frontend
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := protojson.Unmarshal(body, req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}
	} else {
		if err := proto.Unmarshal(body, req); err != nil {
			http.Error(w, "Invalid protobuf payload", http.StatusBadRequest)
			return
		}
	}

	teamsList := req.GetTeams()
	if len(teamsList) == 0 {
		http.Error(w, "At least one team is required", http.StatusBadRequest)
		return
	}

	n := len(teamsList)
	p := nextPowerOfTwo(n)

	// Append BYEs if not power of two
	finalTeams := make([]string, len(teamsList))
	copy(finalTeams, teamsList)
	for len(finalTeams) < p {
		finalTeams = append(finalTeams, "BYE")
	}

	shuffle(finalTeams)

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "DELETE FROM tournaments")
	if err != nil {
		http.Error(w, "Failed to wipe existing tournaments", http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec(ctx, "DELETE FROM teams")
	if err != nil {
		http.Error(w, "Failed to wipe existing teams", http.StatusInternalServerError)
		return
	}

	var tournamentID int
	err = tx.QueryRow(ctx, `
		INSERT INTO tournaments (title, start_time) 
		VALUES ($1, $2) RETURNING tournament_id`,
		req.GetTitle(), req.GetStartTime()).Scan(&tournamentID)
	if err != nil {
		http.Error(w, "Failed to insert tournament", http.StatusInternalServerError)
		return
	}

	teamIDs := make(map[string]int)
	for _, teamName := range finalTeams {
		if _, exists := teamIDs[teamName]; exists {
			continue
		}
		var teamID int
		err = tx.QueryRow(ctx, `
			INSERT INTO teams (team_name) VALUES ($1)
			ON CONFLICT (team_name) DO UPDATE SET team_name = EXCLUDED.team_name
			RETURNING team_id`, teamName).Scan(&teamID)
		if err != nil {
			http.Error(w, "Failed to insert team: "+teamName, http.StatusInternalServerError)
			return
		}
		teamIDs[teamName] = teamID
	}

	k := 0
	for temp := p; temp > 1; temp /= 2 {
		k++
	}

	matchIDs := make(map[string]int)

	for round := 1; round <= k; round++ {
		numMatches := p / (1 << round)
		for pos := 1; pos <= numMatches; pos++ {
			var team1Val, team2Val, team1PrevVal, team2PrevVal *int

			if round == 1 {
				t1Name := finalTeams[2*(pos-1)]
				t2Name := finalTeams[2*(pos-1)+1]
				id1 := teamIDs[t1Name]
				id2 := teamIDs[t2Name]
				team1Val = &id1
				team2Val = &id2
			} else {
				prevPos1 := 2*pos - 1
				prevPos2 := 2*pos
				id1 := matchIDs[mapKey(round-1, prevPos1)]
				id2 := matchIDs[mapKey(round-1, prevPos2)]
				team1PrevVal = &id1
				team2PrevVal = &id2
			}

			var matchID int
			err = tx.QueryRow(ctx, `
				INSERT INTO matches (tournament_id, round_number, match_position, team1_id, team2_id, team1_prev_match_id, team2_prev_match_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING match_id`,
				tournamentID, round, pos, team1Val, team2Val, team1PrevVal, team2PrevVal).Scan(&matchID)
			if err != nil {
				http.Error(w, "Failed to insert match", http.StatusInternalServerError)
				return
			}

			matchIDs[mapKey(round, pos)] = matchID
		}
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	bracketResp, err := h.fetchActiveTournamentBracket(ctx, "", false)
	if err != nil {
		http.Error(w, "Failed to fetch generated bracket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &pb.TournamentUploadResponse{
		Status:         "Tournament generated and uploaded successfully",
		UpdatedBracket: bracketResp,
	}

	// Because the frontend fetch expects JSON, you may want to return JSON here 
	// based on the Accept header, but this leaves the protobuf binary response intact 
	// if your fetch client parses it correctly.
	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// FIX 2: Properly format integers into map keys
func mapKey(round, position int) string {
	return fmt.Sprintf("%d_%d", round, position)
}