package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Format      string // "json" or "text"
	Level       string // "debug", "info", "warn", "error"
	Environment string
}

// SetupLogger configures the global slog logger
func SetupLogger(config LoggerConfig) *slog.Logger {
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if config.Format == "json" || config.Environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// RequestLogger returns a Gin middleware for structured request logging
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// Get request ID if exists
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.Writer.Header().Get("X-Request-ID")
		}

		// Log attributes
		attrs := []slog.Attr{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", clientIP),
		}

		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}
		if requestID != "" {
			attrs = append(attrs, slog.String("request_id", requestID))
		}
		if userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}

		// Get user info from context if authenticated
		if userID := c.GetString("user_id"); userID != "" {
			attrs = append(attrs, slog.String("user_id", userID))
		}

		// Check for errors
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		// Log based on status code
		logArgs := make([]any, len(attrs))
		for i, attr := range attrs {
			logArgs[i] = attr
		}

		switch {
		case status >= 500:
			logger.Error("Server error", logArgs...)
		case status >= 400:
			logger.Warn("Client error", logArgs...)
		case status >= 300:
			logger.Info("Redirect", logArgs...)
		default:
			logger.Info("Request", logArgs...)
		}
	}
}
