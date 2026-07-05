package handler

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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

func mapKey(round, position int) string {
	return fmt.Sprintf("%d_%d", round, position)
}

func (h *Handler) handleSubmitTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	req := &pb.SubmitTeamsRequest{}
	if err := proto.Unmarshal(body, req); err != nil {
		http.Error(w, "Invalid protobuf payload", http.StatusBadRequest)
		return
	}

	teamsList := req.GetTeams()
	if len(teamsList) == 0 {
		http.Error(w, "At least one team is required", http.StatusBadRequest)
		return
	}

	teamListLen := len(teamsList)
	totalTeams := nextPowerOfTwo(teamListLen)

	finalTeams := make([]string, len(teamsList))
	copy(finalTeams, teamsList)
	for len(finalTeams) < totalTeams {
		finalTeams = append(finalTeams, "BYE")
	}

	shuffle(finalTeams)

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "DELETE FROM match_predictions")
	if err != nil {
		http.Error(w, "Failed to wipe match predictions", http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec(ctx, "DELETE FROM matches")
	if err != nil {
		http.Error(w, "Failed to wipe existing matches", http.StatusInternalServerError)
		return
	}
	_, err = tx.Exec(ctx, "DELETE FROM teams")
	if err != nil {
		http.Error(w, "Failed to wipe existing teams", http.StatusInternalServerError)
		return
	}

	teamIDs := make(map[string]int)
	for _, teamName := range finalTeams {
		if _, exists := teamIDs[teamName]; exists {
			continue
		}
		var teamID int
		isBye := strings.ToUpper(teamName) == "BYE"

		err = tx.QueryRow(ctx, `
			INSERT INTO teams (team_name, is_bye) VALUES ($1, $2)
			ON CONFLICT (team_name) DO UPDATE SET team_name = EXCLUDED.team_name, is_bye = EXCLUDED.is_bye
			RETURNING team_id`, teamName, isBye).Scan(&teamID)
		if err != nil {
			http.Error(w, "Failed to insert team: "+teamName, http.StatusInternalServerError)
			return
		}
		teamIDs[teamName] = teamID
	}

	numRounds := 0
	for temp := totalTeams; temp > 1; temp /= 2 {
		numRounds++
	}

	matchIDs := make(map[string]int)
	currentMatchID := 0

	batch := &pgx.Batch{}

	for round := 0; round < numRounds; round++ {
		numMatches := totalTeams / (1 << (round + 1))

		for pos := 0; pos < numMatches; pos++ {
			var team1Val, team2Val, team1PrevVal, team2PrevVal *int

			if round == 0 {
				t1Name := finalTeams[2*pos]
				t2Name := finalTeams[2*pos+1]
				id1 := teamIDs[t1Name]
				id2 := teamIDs[t2Name]
				team1Val = &id1
				team2Val = &id2
			} else {
				prevPos1 := 2 * pos
				prevPos2 := 2*pos + 1
				id1 := matchIDs[mapKey(round-1, prevPos1)]
				id2 := matchIDs[mapKey(round-1, prevPos2)]
				team1PrevVal = &id1
				team2PrevVal = &id2
			}

			batch.Queue(`
				INSERT INTO matches (match_id, round_number, visual_position, team1_id, team2_id, team1_prev_match_id, team2_prev_match_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				currentMatchID, round, pos, team1Val, team2Val, team1PrevVal, team2PrevVal,
			)

			matchIDs[mapKey(round, pos)] = currentMatchID
			currentMatchID++
		}
	}

	br := tx.SendBatch(ctx, batch)

	if err := br.Close(); err != nil {
		http.Error(w, "Failed to execute batch match inserts", http.StatusInternalServerError)
		return
	}

	bracketResp, err := h.fetchActiveTournamentBracket(ctx, "", false)
	if err != nil {
		http.Error(w, "Failed to fetch generated bracket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &pb.SubmitTeamsResponse{
		Status:         "Tournament generated and uploaded successfully",
		UpdatedBracket: bracketResp,
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
