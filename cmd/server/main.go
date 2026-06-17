package main

import (
	"net/http"

	"log"
	"time"
	starter "vercel-go-starter"
	"vercel-go-starter/internal/handler"

	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const dbSchema = `
CREATE TABLE IF NOT EXISTS teams (
    team_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    team_name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tournaments (
    tournament_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS matches (
    match_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tournament_id INT NOT NULL,
    round_number INT NOT NULL, 
    match_position INT NOT NULL, 

    team1_id INT NULL,
    team2_id INT NULL,
    team1_score INT NULL,
    team2_score INT NULL,
    winner_id INT NULL,

    next_match_id INT NULL,
    next_match_slot INT NULL, 

    CONSTRAINT fk_tournament FOREIGN KEY (tournament_id) REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
    CONSTRAINT fk_team1 FOREIGN KEY (team1_id) REFERENCES teams(team_id),
    CONSTRAINT fk_team2 FOREIGN KEY (team2_id) REFERENCES teams(team_id),
    CONSTRAINT fk_winner FOREIGN KEY (winner_id) REFERENCES teams(team_id),
    CONSTRAINT fk_next_match FOREIGN KEY (next_match_id) REFERENCES matches(match_id),
    CONSTRAINT chk_slot CHECK (next_match_slot IN (1, 2)),
    CONSTRAINT uniq_match_pos UNIQUE (tournament_id, round_number, match_position)
);

CREATE TABLE IF NOT EXISTS users (
    user_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_brackets (
    user_bracket_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL,
    tournament_id INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_user_tournament FOREIGN KEY (tournament_id) REFERENCES tournaments(tournament_id) ON DELETE CASCADE,
    CONSTRAINT uniq_user_tournament UNIQUE (user_id, tournament_id)
);

CREATE TABLE IF NOT EXISTS match_predictions (
    prediction_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_bracket_id INT NOT NULL,
    match_id INT NOT NULL,
    predicted_winner_id INT NULL,
    predicted_team1_score INT NULL,
    predicted_team2_score INT NULL,

    CONSTRAINT fk_user_bracket FOREIGN KEY (user_bracket_id) REFERENCES user_brackets(user_bracket_id) ON DELETE CASCADE,
    CONSTRAINT fk_predicted_match FOREIGN KEY (match_id) REFERENCES matches(match_id) ON DELETE CASCADE,
    CONSTRAINT fk_predicted_winner FOREIGN KEY (predicted_winner_id) REFERENCES teams(team_id),
    CONSTRAINT uniq_bracket_match UNIQUE (user_bracket_id, match_id)
);

CREATE INDEX IF NOT EXISTS idx_matches_tournament ON matches(tournament_id);
CREATE INDEX IF NOT EXISTS idx_match_predictions_bracket ON match_predictions(user_bracket_id);
CREATE INDEX IF NOT EXISTS idx_match_predictions_match ON match_predictions(match_id);
`

func main() {
	mux := http.NewServeMux()

	h := handler.New(starter.StaticFiles)
	h.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.ListenAndServe(":"+port, mux)

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	// Get the connection string from the environment variable
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		fmt.Fprintf(os.Stderr, "DATABASE_URL not set\n")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Establish Connection using pgx
	log.Println("Connecting to PostgreSQL...")
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Ping the connection to ensure it is alive
	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Successfully connected to the database.")

	// 3. Execute the Schema DDL Script
	log.Println("Initializing database schema...")
	_, err = conn.Exec(ctx, dbSchema)
	if err != nil {
		log.Fatalf("Failed to create tables: %v\n", err)
	}
	log.Println("Database tables and indexes created successfully!")
}
