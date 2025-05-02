package config

import (
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
    configContent := `
grpc:
  port: 50051
  shutdown_timeout: 10s
log:
  level: debug
`
    tmpfile, err := os.CreateTemp("", "config_test.yaml")
    require.NoError(t, err)
    defer os.Remove(tmpfile.Name())

    _, err = tmpfile.WriteString(configContent)
    require.NoError(t, err)
    require.NoError(t, tmpfile.Close())

    cfg, err := Load(tmpfile.Name())
    require.NoError(t, err)

    assert.Equal(t, 50051, cfg.GRPC.Port)
    assert.Equal(t, 10*time.Second, cfg.GRPC.ShutdownTimeout)
    assert.Equal(t, "debug", cfg.Log.Level)
}

func TestLoadConfig_FileNotExists(t *testing.T) {
    _, err := Load("nonexistent_file.yaml")
    assert.Error(t, err)
    assert.True(t, os.IsNotExist(err), "Should return 'not exists' error")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
    tmpfile, err := os.CreateTemp("", "invalid_config_test.yaml")
    require.NoError(t, err)
    defer os.Remove(tmpfile.Name())

    _, err = tmpfile.WriteString("invalid: yaml: content")
    require.NoError(t, err)
    require.NoError(t, tmpfile.Close())

    _, err = Load(tmpfile.Name())
    assert.Error(t, err)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
    tmpfile, err := os.CreateTemp("", "empty_config_test.yaml")
    require.NoError(t, err)
    defer os.Remove(tmpfile.Name())
    require.NoError(t, tmpfile.Close())

    _, err = Load(tmpfile.Name())
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unmarshal errors")
}

func TestLoadConfig_InvalidDuration(t *testing.T) {
    configContent := `
grpc:
  shutdown_timeout: invalid
`
    tmpfile, err := os.CreateTemp("", "invalid_duration_test.yaml")
    require.NoError(t, err)
    defer os.Remove(tmpfile.Name())

    _, err = tmpfile.WriteString(configContent)
    require.NoError(t, err)
    require.NoError(t, tmpfile.Close())

    _, err = Load(tmpfile.Name())
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "cannot unmarshal")
}