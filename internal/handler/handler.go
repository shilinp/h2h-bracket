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

	mux.HandleFunc("GET /special", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webFS, "special.html")
	})

	mux.HandleFunc("GET /upload", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webFS, "upload.html")
	})

	// Public routes - everything else served from root
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	mux.HandleFunc("GET /api/bracket", h.handleGetTournament)
	mux.HandleFunc("POST /api/bracket", h.handleSubmitBracket)
	mux.HandleFunc("DELETE /api/bracket", h.handleDeleteBracket)
	mux.HandleFunc("POST /api/tournament", h.handleUploadTournament)
}
