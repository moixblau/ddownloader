package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port    string
	DataDir string
	Logger  *slog.Logger
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &Config{
		Port:    port,
		DataDir: dataDir,
		Logger:  logger,
	}
}
