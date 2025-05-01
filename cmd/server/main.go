package main

import (
	"context"
    "fmt"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/StepanErshov/pubsub/config"
    "github.com/StepanErshov/pubsub/internal/service"
    "github.com/StepanErshov/pubsub/pkg/subpub"
)

func main() {
	configureLogger()
	log.Info().Msg("=== SERVER STARTING ===")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	setLogLevel(cfg.Log.Level)

	bus := subpub.NewSubPub()
	defer bus.Close(context.Background())

	grpcServer := grpc.NewServer()
	service.NewPubSubService(bus).Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}
	log.Info().Str("address", lis.Addr().String()).Msg("Server listening on")

	go func() {
        log.Info().Int("port", cfg.GRPC.Port).Msg("Starting gRPC server")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatal().Err(err).Msg("Failed to serve")
        }
    }()

    log.Info().Msg("Server started. Press Ctrl+C to stop")
	
	sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
    
    log.Info().Msg("Shutting down")

	waitForShutdown(grpcServer, cfg.GRPC.ShutdownTimeout)
	log.Info().Str("address", lis.Addr().String()).Msg("Server listening on port")
    log.Info().Msg("=== SERVER READY ===")
}

func configureLogger() {
    output := zerolog.ConsoleWriter{
        Out:        os.Stdout,
        TimeFormat: time.RFC3339,
    }
    log.Logger = zerolog.New(output).With().Timestamp().Logger()
    zerolog.SetGlobalLevel(zerolog.DebugLevel)
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

func setLogLevel(level string) {
    logLevel, err := zerolog.ParseLevel(level)
    if err != nil {
        logLevel = zerolog.InfoLevel 
    }
    zerolog.SetGlobalLevel(logLevel)
}