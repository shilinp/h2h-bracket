package server

import (
	"context"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	"h2h-bracket/server/constants"
	"h2h-bracket/server/proto"

	"github.com/jackc/pgx/v5"
)

// HandleFetchBracket retrieves and maps tournament data based on query criteria.
func (app *App) HandleFetchBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.URL.Query().Get("username")
	isSpecialUser := r.URL.Query().Get("is_special_user") == "true"

	if isSpecialUser {
		username = constants.SpecialUsername
	}

	resp, err := app.fetchActiveTournamentBracket(ctx, username, isSpecialUser)
	if err != nil {
		http.Error(w, "Failed to fetch tournament bracket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	marshaler := protojson.MarshalOptions{
		EmitDefaultValues: true,
	}
	data, err := marshaler.Marshal(resp)

	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// int32Ptr safely copies an int reference over to a raw int32 pointer type.
func int32Ptr(i *int) *int32 {
	if i == nil {
		return nil
	}
	val := int32(*i)
	return &val
}

// fetchActiveTournamentBracket aggregates structure, status, teams, and analytics.
func (app *App) fetchActiveTournamentBracket(ctx context.Context, username string, isSpecialUser bool) (*proto.FetchBracketResponse, error) {
	resp := &proto.FetchBracketResponse{
		Matches:           []*proto.Match{},
		MatchPositions:    make(map[int32]*proto.MatchPosition),
		Predictions:       make(map[int32]int32),
		TeamNames:         make(map[int32]string),
		MasterPredictions: make(map[int32]int32),
	}

	if err := app.loadGlobalLockStatus(ctx, resp); err != nil {
		return nil, err
	}

	if err := app.loadMatchesAndPositions(ctx, resp); err != nil {
		return nil, err
	}

	app.loadTeamNames(ctx, resp)

	_, err := app.resolveUserAndPredictions(ctx, username, resp)
	if err != nil {
		return nil, err
	}

	if err := app.loadMasterDataAndAccuracy(ctx, resp); err != nil {
		return nil, err
	}

	if isSpecialUser || username == constants.SpecialUsername {
		resp.IsLocked = false
	}

	return resp, nil
}

// loadGlobalLockStatus reads the globally shared submission lock settings.
func (app *App) loadGlobalLockStatus(ctx context.Context, resp *proto.FetchBracketResponse) error {
	err := app.DB.QueryRow(ctx, "SELECT is_locked FROM global_settings LIMIT 1").Scan(&resp.IsLocked)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	return nil
}

// loadMatchesAndPositions queries matches to map layout arrays and dictionary references.
func (app *App) loadMatchesAndPositions(ctx context.Context, resp *proto.FetchBracketResponse) error {
	rows, err := app.DB.Query(ctx, `
		SELECT match_id, round_number, visual_position, team1_id, team2_id, team1_prev_match_id, team2_prev_match_id 
		FROM matches ORDER BY round_number, visual_position`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var matchID, roundNumber, visualPosition int
		var team1ID, team2ID, team1PrevID, team2PrevID *int
		if err := rows.Scan(&matchID, &roundNumber, &visualPosition, &team1ID, &team2ID, &team1PrevID, &team2PrevID); err != nil {
			continue
		}

		mID := int32(matchID)
		resp.Matches = append(resp.Matches, &proto.Match{
			MatchId:          mID,
			Team1Id:          int32Ptr(team1ID),
			Team2Id:          int32Ptr(team2ID),
			Team1PrevMatchId: int32Ptr(team1PrevID),
			Team2PrevMatchId: int32Ptr(team2PrevID),
		})

		resp.MatchPositions[mID] = &proto.MatchPosition{
			RoundNumber:    int32(roundNumber),
			VisualPosition: int32(visualPosition),
		}
	}
	return nil
}

// loadTeamNames retrieves all competitive group identities.
func (app *App) loadTeamNames(ctx context.Context, resp *proto.FetchBracketResponse) {
	tRows, err := app.DB.Query(ctx, "SELECT team_id, team_name FROM teams")
	if err != nil {
		return
	}
	defer tRows.Close()

	for tRows.Next() {
		var teamID int
		var teamName string
		if err := tRows.Scan(&teamID, &teamName); err == nil {
			resp.TeamNames[int32(teamID)] = teamName
		}
	}
}

// resolveUserAndPredictions checks or creates a profile and matches user choices.
func (app *App) resolveUserAndPredictions(ctx context.Context, username string, resp *proto.FetchBracketResponse) (int, error) {
	if username == "" {
		return 0, nil
	}

	var userID int
	err := app.DB.QueryRow(ctx, `
		INSERT INTO users (username) VALUES ($1) 
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username 
		RETURNING user_id`, username).Scan(&userID)
	if err != nil {
		return 0, err
	}

	pRows, err := app.DB.Query(ctx, `
		SELECT match_id, predicted_winner_id 
		FROM match_predictions WHERE user_id = $1`, userID)
	if err != nil {
		return userID, err
	}
	defer pRows.Close()

	for pRows.Next() {
		var matchID int
		var predictedWinnerID *int
		pRows.Scan(&matchID, &predictedWinnerID)
		if predictedWinnerID != nil {
			resp.Predictions[int32(matchID)] = int32(*predictedWinnerID)
		}
	}

	if err = pRows.Err(); err != nil {
		return userID, err
	}

	return userID, nil
}

// loadMasterDataAndAccuracy handles specialized comparative analytics scoring.
func (app *App) loadMasterDataAndAccuracy(ctx context.Context, resp *proto.FetchBracketResponse) error {
	var masterID int
	err := app.DB.QueryRow(ctx, "SELECT user_id FROM users WHERE username = $1", constants.SpecialUsername).Scan(&masterID)
	if err != nil {
		return nil
	}

	mRows, err := app.DB.Query(ctx, `
		SELECT match_id, predicted_winner_id 
		FROM match_predictions WHERE user_id = $1`, masterID)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var matchID int
			var predictedWinnerID *int
			mRows.Scan(&matchID, &predictedWinnerID)
			if predictedWinnerID != nil {
				resp.MasterPredictions[int32(matchID)] = int32(*predictedWinnerID)
			}
		}
	}

	// Calculate NCAA scored accuracy if both user and master predictions exist
	if len(resp.Predictions) > 0 && len(resp.MasterPredictions) > 0 {
		var totalPoints float64
		var maxRoundNumber int32

		for matchID, userPick := range resp.Predictions {
			pos, posExists := resp.MatchPositions[matchID]
			if posExists && pos.RoundNumber >= maxRoundNumber {
				maxRoundNumber = pos.RoundNumber
			}
			if masterPick, matchExists := resp.MasterPredictions[matchID]; matchExists {
				if userPick == masterPick {
					if posExists && pos.RoundNumber >= 0 {
						// NCAA scoring: 2^(round - 1)
						points := 1 << (pos.RoundNumber)
						totalPoints += float64(points)
					}
				}
			}

		}
		accuracy := (totalPoints / float64((maxRoundNumber+1)*(1<<maxRoundNumber))) * 100
		resp.Accuracy = &accuracy
	}
	return nil
}
