package server

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/encoding/protojson"

	"h2h-bracket/server/proto"
)

// HandleSubmitTeams acts as the coordinator for processing team inputs and generating a fresh tournament.
func (app *App) HandleSubmitTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseSubmitTeamsRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	finalTeams := seedTeams(req.GetTeams())

	tx, err := app.DB.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	if err := app.clearExistingTournament(ctx, tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	teamIDs, err := app.insertTeams(ctx, tx, finalTeams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.buildAndBatchMatches(ctx, tx, finalTeams, teamIDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Failed to commit database transaction", http.StatusInternalServerError)
		return
	}

	app.respondWithActiveBracket(ctx, w)
}

// parseSubmitTeamsRequest handles low-level I/O reading and Protobuf unmarshaling.
func parseSubmitTeamsRequest(bodyReader io.Reader) (*proto.SubmitTeamsRequest, error) {
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Failed to read request body")
	}

	req := &proto.SubmitTeamsRequest{}
	if err := protojson.Unmarshal(body, req); err != nil {
		return nil, fmt.Errorf("Invalid JSON payload")
	}

	if len(req.GetTeams()) == 0 {
		return nil, fmt.Errorf("At least one team is required")
	}

	return req, nil
}

// seedTeams shuffles the raw team names and balances out the spacing using BYEs to reach a power of two.
func seedTeams(teamsList []string) []string {
	teamListLen := len(teamsList)
	totalTeams := nextPowerOfTwo(teamListLen)

	shuffledTeams := make([]string, teamListLen)
	copy(shuffledTeams, teamsList)
	shuffle(shuffledTeams)

	finalTeams := make([]string, totalTeams)
	numMatchesRound0 := totalTeams / 2
	numByes := totalTeams - teamListLen
	teamIdx := 0

	for pos := 0; pos < numMatchesRound0; pos++ {
		// Distribute BYEs symmetrically across the positions
		if numByes > 0 && (pos < numByes/2 || (numMatchesRound0-1-pos) < (numByes-numByes/2)) {
			finalTeams[2*pos] = shuffledTeams[teamIdx]
			finalTeams[2*pos+1] = "BYE"
			teamIdx++
		} else {
			finalTeams[2*pos] = shuffledTeams[teamIdx]
			teamIdx++
			finalTeams[2*pos+1] = shuffledTeams[teamIdx]
			teamIdx++
		}
	}

	return finalTeams
}

// clearExistingTournament purges the tables to make room for the new bracket run.
func (app *App) clearExistingTournament(ctx context.Context, tx pgx.Tx) error {
	query := `
        TRUNCATE TABLE 
            match_predictions, 
            matches, 
            teams, 
            global_settings 
        RESTART IDENTITY CASCADE;
    `

	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to clear tournament data: %w", err)
	}

	return nil
}

// insertTeams saves unique teams into the database and builds a lookup map linking names to their auto-generated IDs.
func (app *App) insertTeams(ctx context.Context, tx pgx.Tx, finalTeams []string) (map[string]int, error) {
	teamIDs := make(map[string]int)

	for _, teamName := range finalTeams {
		if _, exists := teamIDs[teamName]; exists {
			continue
		}

		var teamID int
		isBye := strings.ToUpper(teamName) == "BYE"

		err := tx.QueryRow(ctx, `
			INSERT INTO teams (team_name, is_bye) VALUES ($1, $2)
			ON CONFLICT (team_name) DO UPDATE SET team_name = EXCLUDED.team_name, is_bye = EXCLUDED.is_bye
			RETURNING team_id`, teamName, isBye).Scan(&teamID)

		if err != nil {
			return nil, fmt.Errorf("Failed to insert team: %s", teamName)
		}
		teamIDs[teamName] = teamID
	}

	return teamIDs, nil
}

// buildAndBatchMatches constructs the multi-round single-elimination structural relationship and pushes it into PGX Batching.
func (app *App) buildAndBatchMatches(ctx context.Context, tx pgx.Tx, finalTeams []string, teamIDs map[string]int) error {
	totalTeams := len(finalTeams)
	numRounds := 0
	for temp := totalTeams; temp > 1; temp /= 2 {
		numRounds++
	}

	matchIDs := make(map[string]int)
	slotTeams := make(map[string]int)
	currentMatchID := 0
	batch := &pgx.Batch{}

	for round := 0; round < numRounds; round++ {
		numMatches := totalTeams / (1 << (round + 1))

		for pos := 0; pos < numMatches; pos++ {
			var team1Val, team2Val, team1PrevVal, team2PrevVal *int

			if round == 0 {
				t1Name, t2Name := finalTeams[2*pos], finalTeams[2*pos+1]
				id1, id2 := teamIDs[t1Name], teamIDs[t2Name]
				team1Val, team2Val = &id1, &id2

				if strings.ToUpper(t1Name) == "BYE" && strings.ToUpper(t2Name) != "BYE" {
					slotTeams[mapKey(round, pos)] = id2
				} else if strings.ToUpper(t1Name) != "BYE" && strings.ToUpper(t2Name) == "BYE" {
					slotTeams[mapKey(round, pos)] = id1
				}
			} else {
				prevPos1, prevPos2 := 2*pos, 2*pos+1
				id1, id2 := matchIDs[mapKey(round-1, prevPos1)], matchIDs[mapKey(round-1, prevPos2)]
				team1PrevVal, team2PrevVal = &id1, &id2

				if advancedTeam1, exists := slotTeams[mapKey(round-1, prevPos1)]; exists {
					team1Val = &advancedTeam1
				}
				if advancedTeam2, exists := slotTeams[mapKey(round-1, prevPos2)]; exists {
					team2Val = &advancedTeam2
				}
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
		return fmt.Errorf("Failed to execute batch match inserts")
	}

	return nil
}

// respondWithActiveBracket pulls the finalized bracket and ships it back over the connection wire.
func (app *App) respondWithActiveBracket(ctx context.Context, w http.ResponseWriter) {
	bracketResp, err := app.fetchActiveTournamentBracket(ctx, "", false)
	if err != nil {
		http.Error(w, "Failed to fetch generated bracket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := &proto.SubmitTeamsResponse{
		Status:         "Tournament generated and uploaded successfully",
		UpdatedBracket: bracketResp,
	}

	data, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

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
