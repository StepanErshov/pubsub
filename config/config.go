package config

import (
    "os"
    "time"

    "gopkg.in/yaml.v3"
)

type Config struct {
    GRPC struct {
        Port            int           `yaml:"port"`
        ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
    } `yaml:"grpc"`
    Log struct {
        Level string `yaml:"level"`
    } `yaml:"log"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}