package config

import (
	"os"
)

type Config struct {
	Port        string
	AppEnv      string
	DatabaseURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")

	return &Config{
		Port:        port,
		AppEnv:      appEnv,
		DatabaseURL: dbURL,
	}
}
