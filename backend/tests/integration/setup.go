package integration

import (
	"context"
	"os"
	"testing"
	"time"
)

// Config holds integration test dependencies.
type Config struct {
	DatabaseURL string
	RedisAddr   string
}

// RequireEnv returns an environment variable or skips the test.
func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("%s not set; skipping integration test", key)
	}
	return value
}

// LoadConfig loads integration config from environment variables.
func LoadConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		DatabaseURL: RequireEnv(t, "DATABASE_URL"),
		RedisAddr:   getenvDefault("REDIS_ADDR", "localhost:6379"),
	}
}

// WithTimeout returns a context with a default timeout for integration tests.
func WithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
