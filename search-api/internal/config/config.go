package config

import (
	"os"
	"strconv"
)

// Config centralizes environment-driven settings for search-api.
type Config struct {
	ServerPort string
	JWTSecret  string

	SolrURL  string
	SolrCore string

	ProductsAPIURL string

	MemcachedAddr   string
	CacheTTLSeconds int
	CacheMaxEntries int64

	RabbitURL      string
	RabbitExchange string
	RabbitQueue    string
}

// Load reads environment variables and applies defaults suitable for local development.
func Load() *Config {
	return &Config{
		ServerPort:      getEnv("PORT", "8082"),
		JWTSecret:       getEnv("JWT_SECRET", "changeme"),
		SolrURL:         getEnv("SOLR_URL", "http://localhost:8983/solr"),
		SolrCore:        getEnv("SOLR_CORE", "products-core"),
		ProductsAPIURL:  getEnv("PRODUCTS_API_URL", "http://localhost:8081"),
		MemcachedAddr:   getEnv("MEMCACHED_ADDR", "localhost:11211"),
		CacheTTLSeconds: getEnvAsInt("CACHE_TTL_SECONDS", 60),
		CacheMaxEntries: getEnvAsInt64("CACHE_MAX_ENTRIES", 1000),
		RabbitURL:       getEnv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/"),
		RabbitExchange:  getEnv("RABBITMQ_EXCHANGE", "products-exchange"),
		RabbitQueue:     getEnv("RABBITMQ_QUEUE", "search-products-queue"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
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

func getEnvAsInt64(key string, def int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return def
	}
	return parsed
}
