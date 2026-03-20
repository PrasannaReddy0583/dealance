package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a new zerolog Logger configured for the given environment.
// In development, it outputs human-readable console format.
// In production, it outputs structured JSON.
func New(env string) zerolog.Logger {
	var w io.Writer

	if env == "development" {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	} else {
		w = os.Stdout
	}

	level := zerolog.InfoLevel
	if env == "development" {
		level = zerolog.DebugLevel
	}

	return zerolog.New(w).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}

// WithService returns a logger with the service name attached.
func WithService(log zerolog.Logger, service string) zerolog.Logger {
	return log.With().Str("service", service).Logger()
}

// WithRequestID returns a logger with the request ID attached.
func WithRequestID(log zerolog.Logger, requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}
