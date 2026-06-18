package main

import (
	"net/http"

	starter "h2h-bracket"
	"h2h-bracket/internal/constants"
	"h2h-bracket/internal/handler"
	"log"
	"time"

	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	// Get the connection string from the environment variable
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatalf("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Establish Connection using pgx
	connPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer connPool.Close()

	// 3. Execute the Schema DDL Script
	log.Println("Initializing database schema...")
	_, err = connPool.Exec(ctx, constants.DBCreationSchema)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}
	log.Println("Database tables and indexes created successfully!")

	mux := http.NewServeMux()

	h := handler.New(starter.StaticFiles, connPool)
	h.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.ListenAndServe(":"+port, mux)
}
