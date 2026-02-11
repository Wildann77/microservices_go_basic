package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog logger
type Logger struct {
	logger zerolog.Logger
}

// contextKey for storing values in context
type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
	ServiceKey contextKey = "service"
)

var defaultLogger *Logger

func init() {
	// Pretty print for development
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(output).With().Timestamp().Logger()
	defaultLogger = &Logger{logger: logger}
}

// New creates a new logger with service name
func New(service string) *Logger {
	return &Logger{
		logger: defaultLogger.logger.With().Str("service", service).Logger(),
	}
}

// WithContext creates a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger

	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		logger = logger.With().Str("trace_id", traceID.(string)).Logger()
	}

	return &Logger{logger: logger}
}

// WithField adds a field to logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{logger: l.logger.With().Interface(key, value).Logger()}
}

// WithError adds error to logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{logger: l.logger.With().Err(err).Logger()}
}

// Debug logs debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Info logs info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Warn logs warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Error logs error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Fatal logs fatal message
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Debugf logs formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// Infof logs formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Warnf logs formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

// Errorf logs formatted error message

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// Fatalf logs formatted fatal message
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal().Msgf(format, v...)
}

// SetTraceID adds trace ID to context
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID gets trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// Global functions for convenience
func Debug(msg string) {
	defaultLogger.Debug(msg)
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Warn(msg string) {
	defaultLogger.Warn(msg)
}

func Error(msg string) {
	defaultLogger.Error(msg)
}

func Fatal(msg string) {
	defaultLogger.Fatal(msg)
}

func WithContext(ctx context.Context) *Logger {
	return defaultLogger.WithContext(ctx)
}
