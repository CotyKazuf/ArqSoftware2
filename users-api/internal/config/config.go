package config

import (
	"os"
	"strconv"
)

// Config centralizes environment-driven settings for the service.
type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	JWTSecret            string
	JWTExpirationMinutes int

	AdminEmail           string
	AdminDefaultPassword string

	ServerPort string
}

// Load reads environment variables and applies sensible defaults for local usage.
func Load() *Config {
	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "users_admin"),
		DBPass:               getEnv("DB_PASS", "users_admin_pass"),
		DBName:               getEnv("DB_NAME", "usersdb"),
		JWTSecret:            getEnv("JWT_SECRET", "changeme"),
		JWTExpirationMinutes: getEnvAsInt("JWT_EXPIRATION", 60),
		AdminEmail:           getEnv("ADMIN_EMAIL", "admin@aromas.com"),
		AdminDefaultPassword: getEnv("ADMIN_DEFAULT_PASSWORD", "admin123"),
		ServerPort:           getEnv("PORT", "8080"),
	}
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getEnvAsInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return parsed
}
