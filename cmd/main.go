package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"vgo-balancer/pkg/config"
	"vgo-balancer/pkg/server"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file.")
	logPath := flag.String("logPath", "./logs/app.log", "Path to store the logs.")
	flag.Parse()

	// Create a logger.
	logger := CreateLogger(*logPath)
	defer logger.Sync()

	// Load the configuration file.
	config, err := config.NewConfig(*configPath)
	if err != nil {
		logger.Fatal("failed to load configuration file", zap.Error(err))
	}
	logger.Info("configuration file loaded successfully")

	// Wait for the signal to shutdown gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Start the Server.
	server := server.NewServer(ctx, logger, config)
	go server.Start()

	<-ctx.Done()
	logger.Info("Load Balancer Shutdown gracefully.")
}
