package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Application settings
	AppEnv    string
	AppPort   string
	LogLevel  string
	LogFormat string

	// Database settings
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis settings
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// Security settings
	JWTSecret  string
	BcryptCost int

	// Rate limiting settings
	RateLimitRequests int
	RateLimitWindow   int
}

func Load() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		AppEnv:    getEnv("APP_ENV", "development"),
		AppPort:   getEnv("APP_PORT", "3000"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "corenglish"),
		DBPassword: getEnv("DB_PASSWORD", "corenglish"),
		DBName:     getEnv("DB_NAME", "corenglish"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsInt("RATE_LIMIT_WINDOW", 60),
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func (c *Config) DatabaseURL() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func (c *Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}
