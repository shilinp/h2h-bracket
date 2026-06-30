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

// handleGetTournament remains exactly the same
func (h *Handler) handleGetTournament(w http.ResponseWriter, r *http.Request) {
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

// Optional pointer helper
func int32Ptr(i *int) *int32 {
	if i == nil {
		return nil
	}
	val := int32(*i)
	return &val
}

func (h *Handler) fetchActiveTournamentBracket(ctx context.Context, username string, isSpecialUser bool) (*pb.FetchMatchupsResponse, error) {
	var tournamentID int
	err := h.DB.QueryRow(ctx, "SELECT tournament_id FROM tournaments LIMIT 1").Scan(&tournamentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &pb.FetchMatchupsResponse{
				Matches:           []*pb.Match{},
				MatchPositions:    make(map[int32]*pb.MatchPosition),
				Predictions:       make(map[int32]int32),
				TeamNames:         make(map[int32]string),
				MasterPredictions: make(map[int32]int32),
			}, nil
		}
		return nil, err
	}

	resp := &pb.FetchMatchupsResponse{
		Matches:           []*pb.Match{},
		MatchPositions:    make(map[int32]*pb.MatchPosition),
		Predictions:       make(map[int32]int32),
		TeamNames:         make(map[int32]string),
		MasterPredictions: make(map[int32]int32),
	}

	// UPDATE: Fetch prev_match_ids instead of next_match_ids
	rows, err := h.DB.Query(ctx, `
		SELECT match_id, round_number, match_position, team1_id, team2_id, team1_prev_match_id, team2_prev_match_id 
		FROM matches WHERE tournament_id = $1 ORDER BY round_number, match_position`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var matchID, roundNumber, matchPosition int
		var team1ID, team2ID, team1PrevID, team2PrevID *int
		err := rows.Scan(&matchID, &roundNumber, &matchPosition, &team1ID, &team2ID, &team1PrevID, &team2PrevID)
		if err != nil {
			log.Println("Error scanning match:", err)
			continue
		}

		mID := int32(matchID)

		// Create the match
		resp.Matches = append(resp.Matches, &pb.Match{
			MatchId:          mID,
			Team1Id:          int32Ptr(team1ID),
			Team2Id:          int32Ptr(team2ID),
			Team1PrevMatchId: int32Ptr(team1PrevID),
			Team2PrevMatchId: int32Ptr(team2PrevID),
		})

		// Populate position mapping
		resp.MatchPositions[mID] = &pb.MatchPosition{
			RoundNumber:    int32(roundNumber),
			VisualPosition: int32(matchPosition),
		}
	}

	// Fetch team names map
	tRows, err := h.DB.Query(ctx, `
		SELECT DISTINCT t.team_id, t.team_name 
		FROM teams t 
		JOIN matches m ON (t.team_id = m.team1_id OR t.team_id = m.team2_id) 
		WHERE m.tournament_id = $1`, tournamentID)
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

	var userBracketID int
	if username != "" {
		h.DB.Exec(ctx, `INSERT INTO users (username) VALUES ($1) ON CONFLICT (username) DO NOTHING`, username)

		isMaster := isSpecialUser || username == constants.SpecialUsername

		err = h.DB.QueryRow(ctx, `
			INSERT INTO user_brackets (user_id, tournament_id, is_master)
			SELECT u.user_id, $2, $3 FROM users u WHERE u.username = $1
			ON CONFLICT (user_id, tournament_id) DO UPDATE SET is_master = user_brackets.is_master OR EXCLUDED.is_master
			RETURNING user_bracket_id`, username, tournamentID, isMaster).Scan(&userBracketID)

		if err == nil {
			var byeTeamID int
			errBye := h.DB.QueryRow(ctx, "SELECT team_id FROM teams WHERE LOWER(team_name) = 'bye' LIMIT 1").Scan(&byeTeamID)
			if errBye == nil {
				// Find all matches containing BYE
				for _, match := range resp.Matches {
					t1 := match.GetTeam1Id()
					t2 := match.GetTeam2Id()
					
					if t1 == int32(byeTeamID) || t2 == int32(byeTeamID) {
						winnerID := t1
						if t1 == int32(byeTeamID) {
							winnerID = t2
						}
						_, _ = h.DB.Exec(ctx, `
							INSERT INTO match_predictions (user_bracket_id, match_id, predicted_winner_id)
							VALUES ($1, $2, $3)
							ON CONFLICT (user_bracket_id, match_id) DO NOTHING`,
							userBracketID, match.GetMatchId(), winnerID)
					}
				}
			}

			// UPDATE: Map predictions directly to the new proto map
			pRows, _ := h.DB.Query(ctx, `
				SELECT match_id, predicted_winner_id 
				FROM match_predictions WHERE user_bracket_id = $1`, userBracketID)
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

	var masterBracketID int
	err = h.DB.QueryRow(ctx, `
		SELECT ub.user_bracket_id FROM user_brackets ub
		JOIN users u ON ub.user_id = u.user_id
		WHERE ub.tournament_id = $1 AND (ub.is_master = TRUE OR u.username = $2)
		LIMIT 1`, tournamentID, constants.SpecialUsername).Scan(&masterBracketID)

	if err == nil {
		resp.IsLocked = userBracketID != 0 && userBracketID != masterBracketID && !isSpecialUser

		// UPDATE: Map master predictions directly to the new proto map
		mRows, _ := h.DB.Query(ctx, `
			SELECT match_id, predicted_winner_id 
			FROM match_predictions WHERE user_bracket_id = $1`, masterBracketID)
		defer mRows.Close()

		for mRows.Next() {
			var matchID int
			var predictedWinnerID *int
			mRows.Scan(&matchID, &predictedWinnerID)
			if predictedWinnerID != nil {
				resp.MasterPredictions[int32(matchID)] = int32(*predictedWinnerID)
			}
		}

		if resp.IsLocked && userBracketID != 0 {
			var total, correct int
			err = h.DB.QueryRow(ctx, `
				SELECT 
					COUNT(*) as total,
					SUM(CASE WHEN up.predicted_winner_id = ap.predicted_winner_id THEN 1 ELSE 0 END) as correct
				FROM match_predictions up
				JOIN match_predictions ap ON up.match_id = ap.match_id
				WHERE up.user_bracket_id = $1 AND ap.user_bracket_id = $2`,
				userBracketID, masterBracketID).Scan(&total, &correct)

			if err == nil && total > 0 {
				acc := (float64(correct) / float64(total)) * 100.0
				resp.Accuracy = &acc
			}
		}
	}

	return resp, nil
}