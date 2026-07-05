package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/bracket", h.handleFetchBracket)
	mux.HandleFunc("POST /api/bracket", h.handleSubmitBracket)
	mux.HandleFunc("DELETE /api/bracket", h.handleDeleteBracket)
	mux.HandleFunc("POST /api/teams", h.handleSubmitTeams)
}
