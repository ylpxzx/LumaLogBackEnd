package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port      string
	DBURL     string
	JWTSecret string
}

func Load() Config {
	dbUser := env("DB_USER", "postgres")
	dbPassword := env("DB_PASSWORD", "794859685")
	dbHost := env("DB_HOST", "localhost")
	dbPort := env("DB_PORT", "5432")
	dbName := env("DB_NAME", "lumalogdb2026")
	dbURL := env("DATABASE_URL", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName))

	return Config{
		Port:      env("PORT", "8080"),
		DBURL:     dbURL,
		JWTSecret: env("JWT_SECRET", "lumalog-dev-secret-change-me"),
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
