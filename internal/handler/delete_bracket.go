package handler

import (
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"

	pb "h2h-bracket/internal/proto"
)

func (h *Handler) handleDeleteBracket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	req := &pb.DeleteBracketRequest{}
	if err := proto.Unmarshal(body, req); err != nil {
		http.Error(w, "Invalid protobuf payload", http.StatusBadRequest)
		return
	}

	username := req.GetUsername()
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	// Remove tournament lookups; map deletion strictly to user_id[cite: 7]
	tag, err := h.DB.Exec(ctx, `
		DELETE FROM match_predictions mp
		USING users u
		WHERE mp.user_id = u.user_id AND u.username = $1`,
		username)

	if err != nil {
		http.Error(w, "Failed to delete bracket", http.StatusInternalServerError)
		return
	}

	if tag.RowsAffected() == 0 {
		http.Error(w, "Bracket not found", http.StatusNotFound)
		return
	}

	resp := &pb.DeleteBracketResponse{Status: "Bracket deleted successfully"}
	w.Header().Set("Content-Type", "application/x-protobuf")

	data, err := proto.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
