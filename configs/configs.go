package configs

import (
	"os"

	"payment-backend/pkg/utils"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() Config {
	_ = godotenv.Load(".env", "../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DatabaseURL: utils.ConnectionURLBuilder(),
		Port:        port,
	}
}
