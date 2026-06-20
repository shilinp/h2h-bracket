package handler

import (
	"encoding/json"
	"net/http"

	"h2h-bracket/internal/constants"
)

func (h *Handler) handleDeleteAdminBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tournamentIDStr := r.URL.Query().Get("tournament_id")

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
