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
	Theme            string
	AuthPassword     string
	Logger           *slog.Logger
}

func (c *Config) IsTransmissionEnabled() bool {
	return c.TransmissionHost != ""
}

func (c *Config) IsAuthEnabled() bool {
	return c.AuthPassword != ""
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

	theme := os.Getenv("THEME")
	if theme == "" {
		theme = "dark"
	}

	transmissionHost := os.Getenv("TRANSMISSION_HOST")
	transmissionUser := os.Getenv("TRANSMISSION_USERNAME")
	transmissionPass := os.Getenv("TRANSMISSION_PASSWORD")

	authPassword := os.Getenv("AUTH_PASSWORD")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &Config{
		Port:             port,
		DataDir:          dataDir,
		TransmissionHost: transmissionHost,
		TransmissionUser: transmissionUser,
		TransmissionPass: transmissionPass,
		Theme:            theme,
		AuthPassword:     authPassword,
		Logger:           logger,
	}
}
