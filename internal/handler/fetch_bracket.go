package handler

import (
	"context"
	"log"
	"net/http"

	"google.golang.org/protobuf/proto"

	"h2h-bracket/internal/constants"
	pb "h2h-bracket/internal/proto"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) handleFetchBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.URL.Query().Get("username")
	isSpecialUser := r.URL.Query().Get("is_special_user") == "true"

	if isSpecialUser {
		username = constants.SpecialUsername
	}

	resp, err := h.fetchActiveTournamentBracket(ctx, username, isSpecialUser)
	if err != nil {
		http.Error(w, "Failed to fetch tournament bracket: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func int32Ptr(i *int) *int32 {
	if i == nil {
		return nil
	}
	val := int32(*i)
	return &val
}

func (h *Handler) fetchActiveTournamentBracket(ctx context.Context, username string, isSpecialUser bool) (*pb.FetchBracketResponse, error) {
	resp := &pb.FetchBracketResponse{
		Matches:           []*pb.Match{},
		MatchPositions:    make(map[int32]*pb.MatchPosition),
		Predictions:       make(map[int32]int32),
		TeamNames:         make(map[int32]string),
		MasterPredictions: make(map[int32]int32),
	}

	err := h.DB.QueryRow(ctx, "SELECT is_locked FROM global_settings LIMIT 1").Scan(&resp.IsLocked)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	rows, err := h.DB.Query(ctx, `
		SELECT match_id, round_number, visual_position, team1_id, team2_id, team1_prev_match_id, team2_prev_match_id 
		FROM matches ORDER BY round_number, visual_position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var matchID, roundNumber, visualPosition int
		var team1ID, team2ID, team1PrevID, team2PrevID *int
		err := rows.Scan(&matchID, &roundNumber, &visualPosition, &team1ID, &team2ID, &team1PrevID, &team2PrevID)
		if err != nil {
			log.Println("Error scanning match:", err)
			continue
		}

		mID := int32(matchID)

		resp.Matches = append(resp.Matches, &pb.Match{
			MatchId:          mID,
			Team1Id:          int32Ptr(team1ID),
			Team2Id:          int32Ptr(team2ID),
			Team1PrevMatchId: int32Ptr(team1PrevID),
			Team2PrevMatchId: int32Ptr(team2PrevID),
		})

		resp.MatchPositions[mID] = &pb.MatchPosition{
			RoundNumber:    int32(roundNumber),
			VisualPosition: int32(visualPosition),
		}
	}

	tRows, err := h.DB.Query(ctx, "SELECT team_id, team_name FROM teams")
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var teamID int
			var teamName string
			if err := tRows.Scan(&teamID, &teamName); err == nil {
				resp.TeamNames[int32(teamID)] = teamName
			}
		}
	}

	var userID int
	if username != "" {
		err = h.DB.QueryRow(ctx, `
			INSERT INTO users (username) VALUES ($1) 
			ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username 
			RETURNING user_id`, username).Scan(&userID)

		if err == nil {
			var byeTeamID int
			errBye := h.DB.QueryRow(ctx, "SELECT team_id FROM teams WHERE is_bye = TRUE LIMIT 1").Scan(&byeTeamID)
			if errBye == nil {
				for _, match := range resp.Matches {
					t1 := match.GetTeam1Id()
					t2 := match.GetTeam2Id()

					if t1 == int32(byeTeamID) || t2 == int32(byeTeamID) {
						winnerID := t1
						if t1 == int32(byeTeamID) {
							winnerID = t2
						}
						_, _ = h.DB.Exec(ctx, `
							INSERT INTO match_predictions (user_id, match_id, predicted_winner_id)
							VALUES ($1, $2, $3)
							ON CONFLICT (user_id, match_id) DO NOTHING`,
							userID, match.GetMatchId(), winnerID)
					}
				}
			}

			pRows, _ := h.DB.Query(ctx, `
				SELECT match_id, predicted_winner_id 
				FROM match_predictions WHERE user_id = $1`, userID)
			defer pRows.Close()

			for pRows.Next() {
				var matchID int
				var predictedWinnerID *int
				pRows.Scan(&matchID, &predictedWinnerID)
				if predictedWinnerID != nil {
					resp.Predictions[int32(matchID)] = int32(*predictedWinnerID)
				}
			}
		}
	}

	var masterID int
	err = h.DB.QueryRow(ctx, "SELECT user_id FROM users WHERE username = $1", constants.SpecialUsername).Scan(&masterID)

	if err == nil {
		mRows, _ := h.DB.Query(ctx, `
			SELECT match_id, predicted_winner_id 
			FROM match_predictions WHERE user_id = $1`, masterID)
		defer mRows.Close()

		for mRows.Next() {
			var matchID int
			var predictedWinnerID *int
			mRows.Scan(&matchID, &predictedWinnerID)
			if predictedWinnerID != nil {
				resp.MasterPredictions[int32(matchID)] = int32(*predictedWinnerID)
			}
		}

		isMasterViewing := isSpecialUser || username == constants.SpecialUsername

		if resp.IsLocked && userID != 0 && !isMasterViewing {
			var total, correct int
			err = h.DB.QueryRow(ctx, `
				SELECT 
					COUNT(*) as total,
					SUM(CASE WHEN up.predicted_winner_id = ap.predicted_winner_id THEN 1 ELSE 0 END) as correct
				FROM match_predictions up
				JOIN match_predictions ap ON up.match_id = ap.match_id
				WHERE up.user_id = $1 AND ap.user_id = $2`,
				userID, masterID).Scan(&total, &correct)

			if err == nil && total > 0 {
				acc := (float64(correct) / float64(total)) * 100.0
				resp.Accuracy = &acc
			}
		}
	}

	if isSpecialUser || username == constants.SpecialUsername {
		resp.IsLocked = false
	}

	return resp, nil
}
