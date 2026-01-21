package config

import (
	"log/slog"
	"os"
)

type Config struct {
	Port             string
	DataDir          string
	TransmissionHost string
	TransmissionUser string
	TransmissionPass string
	Logger           *slog.Logger
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

	transmissionHost := os.Getenv("TRANSMISSION_HOST")
	transmissionUser := os.Getenv("TRANSMISSION_USERNAME")
	transmissionPass := os.Getenv("TRANSMISSION_PASSWORD")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &Config{
		Port:             port,
		DataDir:          dataDir,
		TransmissionHost: transmissionHost,
		TransmissionUser: transmissionUser,
		TransmissionPass: transmissionPass,
		Logger:           logger,
	}
}
