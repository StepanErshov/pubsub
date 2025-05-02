package main

import (
    "os"
    "syscall"
    "testing"
    "time"
	
	"google.golang.org/grpc"
	"github.com/rs/zerolog"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConfigureLogger(t *testing.T) {
    configureLogger()
    assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

func TestWaitForShutdown(t *testing.T) {
    server := grpc.NewServer()
    timeout := 100 * time.Millisecond

    go func() {
        time.Sleep(10 * time.Millisecond)
        proc, err := os.FindProcess(os.Getpid())
        require.NoError(t, err)
        proc.Signal(syscall.SIGINT)
    }()

    start := time.Now()
    waitForShutdown(server, timeout)
    elapsed := time.Since(start)

    assert.True(t, elapsed < timeout, "Should shutdown before timeout")
}

func TestMainFunction(t *testing.T) {

    configContent := `
grpc:
  port: 50051
  shutdown_timeout: 1s
log:
  level: debug
`
    tmpfile, err := os.CreateTemp("", "server_config_test.yaml")
    require.NoError(t, err)
    defer os.Remove(tmpfile.Name())

    _, err = tmpfile.WriteString(configContent)
    require.NoError(t, err)
    require.NoError(t, tmpfile.Close())

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"cmd", tmpfile.Name()}

    done := make(chan struct{})
    go func() {
        main()
        close(done)
    }()

    time.Sleep(100 * time.Millisecond)
    proc, err := os.FindProcess(os.Getpid())
    require.NoError(t, err)
    proc.Signal(syscall.SIGINT)

    select {
    case <-done:
    case <-time.After(2 * time.Second):
        t.Fatal("Main function didn't exit on signal")
    }
}