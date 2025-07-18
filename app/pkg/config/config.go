package config

import (
	"kentech-project/pkg/logger"
	"os"

	"github.com/joho/godotenv"
)

// Config Using godotenv to load environment variables from a .env file if it exists.
// the choice of using godotenv is to allow for easy local development and testing keeping the simplicity of the application in mind.
// to a more complex configuration (such as per environment and with a lot of options), I would choose to use env.yaml with viper or similar libraries.)
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	WalletURL   string
}

func Load() (*Config, error) {
	log := logger.New()
	if err := godotenv.Load(); err != nil {
		log.Errorf("Failed to load .env file: %v", err)
	} else {
		log.Info("Successfully loaded environment variables from .env file")
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/kentech_db?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "defaultsecretkey"),
		WalletURL:   getEnv("WALLET_URL", "http://localhost:9090"),
	}

	log.Debugf("Config loaded: %+v", cfg)
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	log := logger.New()
	if value := os.Getenv(key); value != "" {
		log.Infof("Environment variable %s loaded: %s", key, value)
		return value
	}
	log.Warnf("Environment variables not set, using default value for %s: %s", key, defaultValue)
	return defaultValue
}
