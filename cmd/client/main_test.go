package main

import (
    "context"
    "testing"
	"time"

    "github.com/stretchr/testify/assert"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func TestMainFunction(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    conn, err := grpc.DialContext(ctx, "localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        t.Skip("gRPC server not available, skipping test")
    }
    defer conn.Close()

    assert.NotNil(t, conn)
}