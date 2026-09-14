package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret string
	Port      string
	CG_KEY    string
}

func Load() *Config {
	// Try to load .env from current directory
	err := godotenv.Load()
	if err != nil {
		// Try to load from parent directory (project root)
		err = godotenv.Load("../.env")
		if err != nil {
			log.Println("Warning: .env file not found, using environment variables")
		}
	}

	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "default-secret-change-me"),
		Port:      getEnv("PORT", "8080"),
		CG_KEY:    getEnv("COIN_GECKO_API_KEY", "some bullsht"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
