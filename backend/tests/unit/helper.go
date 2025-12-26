package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// NewTestContext returns a background context for unit tests.
func NewTestContext(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

// MustParseUUID parses a UUID or fails the test.
func MustParseUUID(t *testing.T, value string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("invalid UUID %q: %v", value, err)
	}
	return parsed
}

// FixedClock returns a deterministic clock function for tests.
func FixedClock(ts time.Time) func() time.Time {
	return func() time.Time {
		return ts
	}
}

// RequireNoError fails the test on unexpected errors.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
