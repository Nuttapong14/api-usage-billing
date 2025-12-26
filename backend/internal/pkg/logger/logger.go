package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// New constructs a zerolog logger with optional console output.
func New(level string, pretty bool) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	lvl := zerolog.InfoLevel
	if parsed, err := zerolog.ParseLevel(strings.ToLower(level)); err == nil {
		lvl = parsed
	}

	var out io.Writer = os.Stdout
	if pretty {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	return zerolog.New(out).Level(lvl).With().Timestamp().Logger()
}
