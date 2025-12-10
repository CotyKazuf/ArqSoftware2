package config

import (
	"os"
)

// Config centralizes environment variables for products-api.
type Config struct {
	MongoURI        string
	MongoDB         string
	JWTSecret       string
	UsersAPIBaseURL string

	RabbitMQURL      string
	RabbitMQExchange string

	ServerPort string
}

// Load builds a Config with defaults suitable for local development.
func Load() *Config {
	return &Config{
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:          getEnv("MONGO_DB_NAME", "productsdb"),
		JWTSecret:        getEnv("JWT_SECRET", "changeme"),
		UsersAPIBaseURL:  getEnv("USERS_API_URL", "http://localhost:8080"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "products-exchange"),
		ServerPort:       getEnv("PORT", "8081"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
