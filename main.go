package main

import (
	"log"
	"price-comparison-tool/internal/api"
	"price-comparison-tool/internal/config"
)

func main() {
	cfg := config.Load()
	
	server := api.NewServer(cfg)
	
	log.Printf("Starting server on port %s", cfg.Port)
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}