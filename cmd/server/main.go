package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"pubsub/config"
	"pubsub/internal/service"
	"pubsub/pkg/subpub"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Setup logger
	configureLogger(cfg)

	// Create pubsub bus
	bus := subpub.NewSubPub()
	defer bus.Close(context.Background())

	// Create gRPC server
	grpcServer := grpc.NewServer()
	service.NewPubSubService(bus).Register(grpcServer)

	// Start server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	go func() {
		log.Info().Int("port", cfg.GRPC.Port).Msg("Starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve")
		}
	}()

	// Graceful shutdown
	waitForShutdown(grpcServer, cfg.GRPC.ShutdownTimeout)
}

func configureLogger(cfg *config.Config) {
	// Configure zerolog based on config
	// ...
}

func waitForShutdown(server *grpc.Server, timeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info().Msg("Server stopped gracefully")
	case <-ctx.Done():
		server.Stop()
		log.Warn().Msg("Server forced to stop")
	}
}