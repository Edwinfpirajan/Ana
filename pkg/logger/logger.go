// Package logger provides structured logging for AnaStreamer
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the global logger instance
var Logger zerolog.Logger

// Init initializes the logger with the specified level
func Init(level string, writer io.Writer) {
	if writer == nil {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	lvl := parseLevel(level)
	zerolog.SetGlobalLevel(lvl)

	Logger = zerolog.New(writer).
		With().
		Timestamp().
		Caller().
		Logger()
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	Logger.Debug().Msgf(format, v...)
}

// Info logs an info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	Logger.Info().Msgf(format, v...)
}

// Warn logs a warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, v ...interface{}) {
	Logger.Warn().Msgf(format, v...)
}

// Error logs an error message
func Error(msg string, err error) {
	Logger.Error().Err(err).Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	Logger.Error().Msgf(format, v...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error) {
	Logger.Fatal().Err(err).Msg(msg)
}

// WithField returns a logger with an additional field
func WithField(key string, value interface{}) zerolog.Logger {
	return Logger.With().Interface(key, value).Logger()
}

// WithFields returns a logger with additional fields
func WithFields(fields map[string]interface{}) zerolog.Logger {
	ctx := Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}

// Component returns a logger for a specific component
func Component(name string) zerolog.Logger {
	return Logger.With().Str("component", name).Logger()
}
