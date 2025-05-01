package config

import (
	"time"
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