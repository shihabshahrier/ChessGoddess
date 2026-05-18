package main

import (
	"context"
	"log"
	"os"

	"github.com/chessgoddess/chesslens/internal/auth"
	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/database"
	"github.com/chessgoddess/chesslens/internal/server"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	authService := auth.NewService(cfg)

	srv := server.New(cfg, db, authService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ChessLens server starting on port %s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
