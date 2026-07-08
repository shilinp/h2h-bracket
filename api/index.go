package handler

import (
	"context"
	"h2h-bracket/server"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	globalMux   *http.ServeMux
	connPool    *pgxpool.Pool
	appInstance *server.App
	initOnce    sync.Once
)

func setupEnvironment() {
	globalMux = http.NewServeMux()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Println("CRITICAL ERROR: DATABASE_URL not set in Vercel environment")
		return
	}

	var err error
	connPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Printf("CRITICAL ERROR: Unable to connect to database: %v\n", err)
		return
	}

	appInstance = &server.App{DB: connPool}
	globalMux.HandleFunc("GET /api/bracket", appInstance.HandleFetchBracket)
	globalMux.HandleFunc("POST /api/bracket", appInstance.HandleSubmitBracket)
	globalMux.HandleFunc("DELETE /api/bracket", appInstance.HandleDeleteBracket)
	globalMux.HandleFunc("POST /api/teams", appInstance.HandleSubmitTeams)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(setupEnvironment)

	if globalMux == nil {
		http.Error(w, "Internal Server Error: Application failed to initialize", http.StatusInternalServerError)
		return
	}

	globalMux.ServeHTTP(w, r)
}
