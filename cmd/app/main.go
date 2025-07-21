package main

import (
	"os"
	"os/signal"
	"syscall"

	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/server"
)

func init() {
	logger.InitLogger()
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log := logger.Logger

	app, err := server.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating server")
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down...")
		if err := app.Shutdown(); err != nil {
			log.Fatal().Err(err).Msg("Server shutdown failed")
		}
		log.Println("Server gracefully stopped")
		os.Exit(0)
	}()

	// Start the server
	log.Printf("Starting server on port %s", cfg.DB.Port)
	if err := app.Start(cfg.DB.Port); err != nil {
		log.Fatal().Err(err).Msg("Error starting server")
	}
}
