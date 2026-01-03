package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Level   string // "error", "warning"
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", strings.ToUpper(e.Level), e.Field, e.Message)
}

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Security   SecurityConfig
	RateLimit  RateLimitConfig
	Features   FeatureConfig
	Validation ValidationResult
}

// ValidationResult holds the config validation summary
type ValidationResult struct {
	Errors   int
	Warnings int
}

type ServerConfig struct {
	Port         string
	Env          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

type SecurityConfig struct {
	AllowedOrigins     []string
	MaxLoginAttempts   int
	LockoutDuration    time.Duration
	MinPasswordLength  int
	BcryptCost         int
	RequireSpecialChar bool
	MaxUploadSizeMB    int
}

type RateLimitConfig struct {
	Enabled         bool
	GlobalPerMin    int
	GlobalWindowSec int
	AuthPerMin      int
	AuthWindowSec   int
	UserPerMin      int
}

type FeatureConfig struct {
	EnableSwagger  bool
	EnableAuditLog bool
	DebugMode      bool
}

// Placeholder patterns to detect insecure default values
var placeholderPatterns = []string{
	"your-", "change-me", "replace-this", "xxx", "placeholder",
	"example", "default", "secret", "password123", "admin123", "test123", "12345",
}

// Weak secrets that should never be used
var weakSecrets = []string{
	"secret", "password", "jwt_secret", "mysecret", "changeme", "qwerty", "abc123",
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	env := getEnv("ENV", "development")
	isProduction := env == "production"

	var errors []ValidationError
	var warnings []ValidationError

	// ============================================
	// SERVER VALIDATION
	// ============================================
	port := getEnv("PORT", "8080")
	if err := validatePort(port); err != nil {
		errors = append(errors, ValidationError{Field: "PORT", Message: err.Error(), Level: "error"})
	}

	if !isValidEnv(env) {
		warnings = append(warnings, ValidationError{
			Field: "ENV", Message: fmt.Sprintf("Unknown environment '%s', expected: development, staging, production", env), Level: "warning",
		})
	}

	readTimeoutSec := getEnvInt("SERVER_READ_TIMEOUT_SEC", 15)
	writeTimeoutSec := getEnvInt("SERVER_WRITE_TIMEOUT_SEC", 15)
	idleTimeoutSec := getEnvInt("SERVER_IDLE_TIMEOUT_SEC", 60)

	// ============================================
	// DATABASE VALIDATION
	// ============================================
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "bukuo_db")
	dbMaxConns := getEnvInt("DB_MAX_CONNS", 25)
	dbMinConns := getEnvInt("DB_MIN_CONNS", 5)
	dbMaxConnLifetimeMin := getEnvInt("DB_MAX_CONN_LIFETIME_MIN", 60)
	dbMaxConnIdleTimeMin := getEnvInt("DB_MAX_CONN_IDLE_TIME_MIN", 30)

	if err := validatePort(dbPort); err != nil {
		errors = append(errors, ValidationError{Field: "DB_PORT", Message: err.Error(), Level: "error"})
	}

	if dbHost == "" {
		errors = append(errors, ValidationError{Field: "DB_HOST", Message: "Database host cannot be empty", Level: "error"})
	}

	if dbName == "" {
		errors = append(errors, ValidationError{Field: "DB_NAME", Message: "Database name cannot be empty", Level: "error"})
	}

	if dbMaxConns < 1 {
		errors = append(errors, ValidationError{Field: "DB_MAX_CONNS", Message: "Max connections must be at least 1", Level: "error"})
	}

	if dbMinConns > dbMaxConns {
		errors = append(errors, ValidationError{Field: "DB_MIN_CONNS", Message: "Min connections cannot exceed max connections", Level: "error"})
	}

	if isProduction {
		if dbPassword == "" {
			errors = append(errors, ValidationError{Field: "DB_PASSWORD", Message: "Database password is required in production", Level: "error"})
		}
		if dbUser == "postgres" {
			warnings = append(warnings, ValidationError{Field: "DB_USER", Message: "Using default 'postgres' user in production is not recommended", Level: "warning"})
		}
		if dbHost == "localhost" || dbHost == "127.0.0.1" {
			warnings = append(warnings, ValidationError{Field: "DB_HOST", Message: "Using localhost in production - ensure this is intentional", Level: "warning"})
		}
	}

	// ============================================
	// JWT VALIDATION
	// ============================================
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtExpiryHours := getEnvInt("JWT_EXPIRY_HOURS", 24)
	jwtRefreshExpiryDays := getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)

	if jwtSecret == "" {
		if isProduction {
			errors = append(errors, ValidationError{Field: "JWT_SECRET", Message: "JWT_SECRET is required in production", Level: "error"})
		} else {
			jwtSecret = "dev-only-secret-not-for-production-use"
			warnings = append(warnings, ValidationError{Field: "JWT_SECRET", Message: "Using default JWT secret - set JWT_SECRET for production", Level: "warning"})
		}
	} else {
		if isPlaceholder(jwtSecret) {
			if isProduction {
				errors = append(errors, ValidationError{Field: "JWT_SECRET", Message: "JWT_SECRET appears to be a placeholder value", Level: "error"})
			} else {
				warnings = append(warnings, ValidationError{Field: "JWT_SECRET", Message: "JWT_SECRET appears to be a placeholder value", Level: "warning"})
			}
		}
		if isWeakSecret(jwtSecret) {
			if isProduction {
				errors = append(errors, ValidationError{Field: "JWT_SECRET", Message: "JWT_SECRET is too weak", Level: "error"})
			} else {
				warnings = append(warnings, ValidationError{Field: "JWT_SECRET", Message: "JWT_SECRET is weak - consider using a stronger secret", Level: "warning"})
			}
		}
		if len(jwtSecret) < 32 {
			if isProduction {
				errors = append(errors, ValidationError{Field: "JWT_SECRET", Message: fmt.Sprintf("JWT_SECRET is too short (%d chars) - minimum 32 required", len(jwtSecret)), Level: "error"})
			} else {
				warnings = append(warnings, ValidationError{Field: "JWT_SECRET", Message: fmt.Sprintf("JWT_SECRET is short (%d chars) - recommend at least 32", len(jwtSecret)), Level: "warning"})
			}
		}
	}

	if jwtExpiryHours < 1 {
		errors = append(errors, ValidationError{Field: "JWT_EXPIRY_HOURS", Message: "JWT expiry must be at least 1 hour", Level: "error"})
	}
	if jwtExpiryHours > 720 {
		warnings = append(warnings, ValidationError{Field: "JWT_EXPIRY_HOURS", Message: fmt.Sprintf("JWT expiry of %d hours is very long", jwtExpiryHours), Level: "warning"})
	}

	// ============================================
	// SECURITY VALIDATION
	// ============================================
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	allowedOrigins := parseOrigins(originsStr)

	if len(allowedOrigins) == 0 {
		errors = append(errors, ValidationError{Field: "ALLOWED_ORIGINS", Message: "At least one allowed origin is required", Level: "error"})
	}

	for _, origin := range allowedOrigins {
		if origin == "*" {
			if isProduction {
				errors = append(errors, ValidationError{Field: "ALLOWED_ORIGINS", Message: "Wildcard '*' not allowed in production", Level: "error"})
			} else {
				warnings = append(warnings, ValidationError{Field: "ALLOWED_ORIGINS", Message: "Wildcard '*' is insecure", Level: "warning"})
			}
		} else if !isValidOriginURL(origin) {
			warnings = append(warnings, ValidationError{Field: "ALLOWED_ORIGINS", Message: fmt.Sprintf("Invalid origin URL: %s", origin), Level: "warning"})
		}
	}

	if isProduction {
		for _, origin := range allowedOrigins {
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				warnings = append(warnings, ValidationError{Field: "ALLOWED_ORIGINS", Message: fmt.Sprintf("Localhost origin '%s' in production", origin), Level: "warning"})
			}
		}
	}

	maxLoginAttempts := getEnvInt("MAX_LOGIN_ATTEMPTS", 5)
	lockoutDurationMin := getEnvInt("LOCKOUT_DURATION_MIN", 15)
	minPasswordLength := getEnvInt("MIN_PASSWORD_LENGTH", 8)
	bcryptCost := getEnvInt("BCRYPT_COST", 10)
	requireSpecialChar := getEnvBool("REQUIRE_SPECIAL_CHAR", false)
	maxUploadSizeMB := getEnvInt("MAX_UPLOAD_SIZE_MB", 10)

	if maxLoginAttempts < 1 {
		errors = append(errors, ValidationError{Field: "MAX_LOGIN_ATTEMPTS", Message: "Must be at least 1", Level: "error"})
	}
	if lockoutDurationMin < 1 {
		errors = append(errors, ValidationError{Field: "LOCKOUT_DURATION_MIN", Message: "Must be at least 1 minute", Level: "error"})
	}
	if minPasswordLength < 6 {
		errors = append(errors, ValidationError{Field: "MIN_PASSWORD_LENGTH", Message: "Must be at least 6", Level: "error"})
	}
	if bcryptCost < 4 || bcryptCost > 31 {
		errors = append(errors, ValidationError{Field: "BCRYPT_COST", Message: "Must be between 4 and 31", Level: "error"})
	}
	if bcryptCost < 10 && isProduction {
		warnings = append(warnings, ValidationError{Field: "BCRYPT_COST", Message: "Cost 10+ recommended for production", Level: "warning"})
	}
	if maxUploadSizeMB < 1 {
		errors = append(errors, ValidationError{Field: "MAX_UPLOAD_SIZE_MB", Message: "Must be at least 1 MB", Level: "error"})
	}

	// ============================================
	// RATE LIMIT VALIDATION
	// ============================================
	rateLimitEnabled := getEnvBool("RATE_LIMIT_ENABLED", true)
	rateLimitPerMin := getEnvInt("RATE_LIMIT_PER_MIN", 100)
	rateLimitWindowSec := getEnvInt("RATE_LIMIT_WINDOW_SEC", 60)
	authRateLimitPerMin := getEnvInt("AUTH_RATE_LIMIT_PER_MIN", 10)
	authRateLimitWindowSec := getEnvInt("AUTH_RATE_LIMIT_WINDOW_SEC", 60)
	userRateLimitPerMin := getEnvInt("USER_RATE_LIMIT_PER_MIN", 200)

	if rateLimitPerMin < 1 {
		errors = append(errors, ValidationError{Field: "RATE_LIMIT_PER_MIN", Message: "Must be at least 1", Level: "error"})
	}
	if rateLimitPerMin > 10000 {
		warnings = append(warnings, ValidationError{Field: "RATE_LIMIT_PER_MIN", Message: fmt.Sprintf("Rate limit of %d/min is very high", rateLimitPerMin), Level: "warning"})
	}
	if authRateLimitPerMin > rateLimitPerMin {
		warnings = append(warnings, ValidationError{Field: "AUTH_RATE_LIMIT_PER_MIN", Message: "Auth rate limit should be stricter than global", Level: "warning"})
	}

	// ============================================
	// FEATURE FLAGS
	// ============================================
	enableSwagger := getEnvBool("ENABLE_SWAGGER", true)
	enableAuditLog := getEnvBool("ENABLE_AUDIT_LOG", true)
	debugMode := getEnvBool("DEBUG_MODE", false)

	if debugMode && isProduction {
		warnings = append(warnings, ValidationError{Field: "DEBUG_MODE", Message: "Debug mode should be disabled in production", Level: "warning"})
	}
	if !enableAuditLog && isProduction {
		warnings = append(warnings, ValidationError{Field: "ENABLE_AUDIT_LOG", Message: "Audit logging should be enabled in production", Level: "warning"})
	}

	// ============================================
	// PRINT VALIDATION RESULTS
	// ============================================
	printValidationResults(errors, warnings, isProduction)

	if len(errors) > 0 && isProduction {
		log.Fatal("❌ Configuration validation failed - fix errors above before starting in production mode")
	}

	return &Config{
		Server: ServerConfig{
			Port:         port,
			Env:          env,
			ReadTimeout:  time.Duration(readTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(writeTimeoutSec) * time.Second,
			IdleTimeout:  time.Duration(idleTimeoutSec) * time.Second,
		},
		Database: DatabaseConfig{
			Host:            dbHost,
			Port:            dbPort,
			User:            dbUser,
			Password:        dbPassword,
			Name:            dbName,
			MaxConns:        dbMaxConns,
			MinConns:        dbMinConns,
			MaxConnLifetime: time.Duration(dbMaxConnLifetimeMin) * time.Minute,
			MaxConnIdleTime: time.Duration(dbMaxConnIdleTimeMin) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:        jwtSecret,
			Expiry:        time.Duration(jwtExpiryHours) * time.Hour,
			RefreshExpiry: time.Duration(jwtRefreshExpiryDays) * 24 * time.Hour,
		},
		Security: SecurityConfig{
			AllowedOrigins:     allowedOrigins,
			MaxLoginAttempts:   maxLoginAttempts,
			LockoutDuration:    time.Duration(lockoutDurationMin) * time.Minute,
			MinPasswordLength:  minPasswordLength,
			BcryptCost:         bcryptCost,
			RequireSpecialChar: requireSpecialChar,
			MaxUploadSizeMB:    maxUploadSizeMB,
		},
		RateLimit: RateLimitConfig{
			Enabled:         rateLimitEnabled,
			GlobalPerMin:    rateLimitPerMin,
			GlobalWindowSec: rateLimitWindowSec,
			AuthPerMin:      authRateLimitPerMin,
			AuthWindowSec:   authRateLimitWindowSec,
			UserPerMin:      userRateLimitPerMin,
		},
		Features: FeatureConfig{
			EnableSwagger:  enableSwagger,
			EnableAuditLog: enableAuditLog,
			DebugMode:      debugMode,
		},
		Validation: ValidationResult{
			Errors:   len(errors),
			Warnings: len(warnings),
		},
	}
}

// ============================================
// VALIDATION HELPERS
// ============================================

func validatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", p)
	}
	if p < 1024 {
		return fmt.Errorf("port %d requires root privileges (use 1024+)", p)
	}
	return nil
}

func isValidEnv(env string) bool {
	validEnvs := []string{"development", "staging", "production", "test"}
	for _, v := range validEnvs {
		if env == v {
			return true
		}
	}
	return false
}

func isPlaceholder(value string) bool {
	lower := strings.ToLower(value)
	for _, pattern := range placeholderPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func isWeakSecret(secret string) bool {
	lower := strings.ToLower(secret)
	for _, weak := range weakSecrets {
		if lower == weak {
			return true
		}
	}
	return false
}

func parseOrigins(originsStr string) []string {
	var origins []string
	for _, o := range strings.Split(originsStr, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func isValidOriginURL(origin string) bool {
	if origin == "*" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func printValidationResults(errors, warnings []ValidationError, isProduction bool) {
	if len(errors) == 0 && len(warnings) == 0 {
		return
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("📋 CONFIGURATION VALIDATION REPORT")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(errors) > 0 {
		log.Println("")
		log.Println("❌ ERRORS (must fix):")
		for _, e := range errors {
			log.Printf("   • %s: %s", e.Field, e.Message)
		}
	}

	if len(warnings) > 0 {
		log.Println("")
		log.Println("⚠️  WARNINGS (recommended to fix):")
		for _, w := range warnings {
			log.Printf("   • %s: %s", w.Field, w.Message)
		}
	}

	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if isProduction {
		log.Printf("📊 Summary: %d error(s), %d warning(s) in PRODUCTION mode", len(errors), len(warnings))
	} else {
		log.Printf("📊 Summary: %d error(s), %d warning(s) in DEVELOPMENT mode", len(errors), len(warnings))
	}
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// ============================================
// ENV HELPERS
// ============================================

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		lower := strings.ToLower(val)
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return fallback
}
