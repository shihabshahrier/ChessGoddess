package main

import (
	"log"
	"os"

	"github.com/chessgoddess/chesslens/internal/config"
	"github.com/chessgoddess/chesslens/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := server.New(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ChessLens server starting on port %s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
