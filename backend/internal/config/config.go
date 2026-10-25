package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port        string
	AppEnv      string
	DatabaseURL string
}

// loadEnvFile reads a simple .env file and sets variables if not already defined in environment.
func loadEnvFile(paths ...string) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"'`)
				if _, exists := os.LookupEnv(key); !exists {
					_ = os.Setenv(key, val)
				}
			}
		}
		break
	}
}

func Load() *Config {
	// Look for .env in current directory or parent directory
	loadEnvFile(".env", "../.env")

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
