package handler

import (
	"embed"
	"io/fs"
	"net/http"

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
	webFS, _ := fs.Sub(h.assets, "web/dist")
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	mux.HandleFunc("GET /api/bracket", h.handleGetBracket)
	mux.HandleFunc("POST /api/bracket", h.handleSubmitBracket)
	mux.HandleFunc("DELETE /api/admin/bracket", h.handleDeleteAdminBracket)
	mux.HandleFunc("POST /api/admin/tournament/upload", h.handleUploadTournament)
}
