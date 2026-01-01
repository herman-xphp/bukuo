package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type SecurityConfig struct {
	AllowedOrigins    []string
	RateLimitPerMin   int
	MaxLoginAttempts  int
	LockoutDuration   time.Duration
	MinPasswordLength int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	env := getEnv("ENV", "development")

	// JWT Secret is REQUIRED in production
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if env == "production" {
			log.Fatal("FATAL: JWT_SECRET environment variable is required in production")
		}
		// Only allow default in development
		jwtSecret = "dev-only-secret-not-for-production-use"
		log.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET in production!")
	}

	// Parse allowed origins
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	allowedOrigins := strings.Split(originsStr, ",")

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  env,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "bukuo_db"),
		},
		JWT: JWTConfig{
			Secret: jwtSecret,
			Expiry: 24 * time.Hour,
		},
		Security: SecurityConfig{
			AllowedOrigins:    allowedOrigins,
			RateLimitPerMin:   getEnvInt("RATE_LIMIT_PER_MIN", 100),
			MaxLoginAttempts:  getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:   15 * time.Minute,
			MinPasswordLength: getEnvInt("MIN_PASSWORD_LENGTH", 8),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		var result int
		if _, err := os.Stat(val); err == nil {
			return fallback
		}
		if n, err := parseIntSafe(val); err == nil {
			result = n
		} else {
			result = fallback
		}
		return result
	}
	return fallback
}

func parseIntSafe(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
