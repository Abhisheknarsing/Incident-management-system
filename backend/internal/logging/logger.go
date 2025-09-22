package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogLevel represents different log levels
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	level slog.Level
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Duration   *time.Duration         `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	File       string                 `json:"file,omitempty"`
	Line       int                    `json:"line,omitempty"`
}

// Config holds logger configuration
type Config struct {
	Level      LogLevel `json:"level"`
	Format     string   `json:"format"` // "json" or "text"
	Output     string   `json:"output"` // "stdout", "stderr", or file path
	AddSource  bool     `json:"add_source"`
	TimeFormat string   `json:"time_format"`
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      LevelInfo,
		Format:     "json",
		Output:     "stdout",
		AddSource:  true,
		TimeFormat: time.RFC3339,
	}
}

// NewLogger creates a new structured logger
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Determine output writer
	var writer io.Writer
	switch config.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", config.Output, err)
		}
		writer = file
	}

	// Configure slog level
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	case LevelFatal:
		level = slog.LevelError // slog doesn't have fatal, use error
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(config.TimeFormat)),
				}
			}
			return a
		},
	}

	// Create handler based on format
	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	logger := slog.New(handler)
	return &Logger{
		Logger: logger,
		level:  level,
	}, nil
}

// WithContext adds context information to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := []any{}

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// Add user ID if available
	if userID := GetUserID(ctx); userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}

	return &Logger{
		Logger: l.Logger.With(attrs...),
		level:  l.level,
	}
}

// WithComponent adds component information to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("component", component)),
		level:  l.level,
	}
}

// WithOperation adds operation information to the logger
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("operation", operation)),
		level:  l.level,
	}
}

// WithMetadata adds metadata to the logger
func (l *Logger) WithMetadata(metadata map[string]interface{}) *Logger {
	attrs := []any{}
	for key, value := range metadata {
		attrs = append(attrs, slog.Any(key, value))
	}
	return &Logger{
		Logger: l.Logger.With(attrs...),
		level:  l.level,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(slog.LevelDebug, msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(slog.LevelInfo, msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(slog.LevelWarn, msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, args ...interface{}) {
	attrs := make([]interface{}, 0, len(args)+2)
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	attrs = append(attrs, args...)
	l.log(slog.LevelError, msg, attrs...)
}

// ErrorWithStack logs an error message with stack trace
func (l *Logger) ErrorWithStack(msg string, err error, args ...interface{}) {
	attrs := make([]interface{}, 0, len(args)+4)
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	// Add stack trace
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	attrs = append(attrs, slog.String("stack_trace", string(stack[:length])))

	attrs = append(attrs, args...)
	l.log(slog.LevelError, msg, attrs...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, args ...interface{}) {
	l.ErrorWithStack(msg, err, args...)
	os.Exit(1)
}

// LogDuration logs the duration of an operation
func (l *Logger) LogDuration(operation string, start time.Time, args ...interface{}) {
	duration := time.Since(start)
	attrs := make([]interface{}, 0, len(args)+2)
	attrs = append(attrs, slog.String("operation", operation))
	attrs = append(attrs, slog.Duration("duration", duration))
	attrs = append(attrs, args...)

	l.log(slog.LevelInfo, fmt.Sprintf("Operation completed: %s", operation), attrs...)
}

// log is the internal logging method
func (l *Logger) log(level slog.Level, msg string, args ...interface{}) {
	if !l.Logger.Enabled(context.Background(), level) {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if ok {
		args = append(args, slog.String("file", file), slog.Int("line", line))
	}

	l.Logger.Log(context.Background(), level, msg, args...)
}

// Context key types for request and user IDs
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// RequestIDMiddleware is a Gin middleware that adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to Gin context
		c.Set("request_id", requestID)

		// Add to request context
		ctx := WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// LoggingMiddleware is a Gin middleware for request logging
func LoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Create logger with context
		reqLogger := logger.WithContext(c.Request.Context()).WithComponent("http")

		// Log request start
		reqLogger.Info("Request started",
			slog.String("method", method),
			slog.String("path", path),
			slog.String("remote_addr", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		// Log request completion
		duration := time.Since(start)
		status := c.Writer.Status()

		logLevel := slog.LevelInfo
		if status >= 400 && status < 500 {
			logLevel = slog.LevelWarn
		} else if status >= 500 {
			logLevel = slog.LevelError
		}

		reqLogger.Logger.Log(c.Request.Context(), logLevel, "Request completed",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.Int("response_size", c.Writer.Size()),
		)
	}
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// Fallback to default logger
		logger, _ := NewLogger(DefaultConfig())
		return logger
	}
	return globalLogger
}

// Helper functions for global logging
func Debug(msg string, args ...interface{}) {
	GetGlobalLogger().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetGlobalLogger().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetGlobalLogger().Warn(msg, args...)
}

func Error(msg string, err error, args ...interface{}) {
	GetGlobalLogger().Error(msg, err, args...)
}

func ErrorWithStack(msg string, err error, args ...interface{}) {
	GetGlobalLogger().ErrorWithStack(msg, err, args...)
}

func Fatal(msg string, err error, args ...interface{}) {
	GetGlobalLogger().Fatal(msg, err, args...)
}
