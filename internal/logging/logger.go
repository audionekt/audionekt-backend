package logging

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
}

// Config holds logging configuration
type Config struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`    // json, console
	OutputPath string `json:"outputPath"` // stdout, stderr, or file path
}

// New creates a new logger instance
func New(config Config) (*Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Set default values
	if config.Format == "" {
		config.Format = "json"
	}
	if config.OutputPath == "" {
		config.OutputPath = "stdout"
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder
	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create output writer
	var output zapcore.WriteSyncer
	switch config.OutputPath {
	case "stdout":
		output = zapcore.AddSync(os.Stdout)
	case "stderr":
		output = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, output, level)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{Logger: logger}, nil
}

// NewDevelopment creates a logger suitable for development
func NewDevelopment() (*Logger, error) {
	return New(Config{
		Level:      "debug",
		Format:     "console",
		OutputPath: "stdout",
	})
}

// NewProduction creates a logger suitable for production
func NewProduction() (*Logger, error) {
	return New(Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	})
}

// WithRequestID adds request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("request_id", requestID))}
}

// WithUserID adds user ID to the logger context
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("user_id", userID))}
}

// WithOperation adds operation name to the logger context
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{Logger: l.Logger.With(zap.String("operation", operation))}
}

// WithDuration adds duration to the logger context
func (l *Logger) WithDuration(duration time.Duration) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Duration("duration", duration))}
}

// WithError adds error to the logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Error(err))}
}

// WithField adds a custom field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(zap.Any(key, value))}
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{Logger: l.Logger.With(zapFields...)}
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path, userAgent, remoteAddr string, userID *string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("remote_addr", remoteAddr),
	}
	
	if userID != nil {
		fields = append(fields, zap.String("user_id", *userID))
	}
	
	l.Info("HTTP request received", fields...)
}

// LogResponse logs an HTTP response
func (l *Logger) LogResponse(method, path string, statusCode int, duration time.Duration, userID *string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
	}
	
	if userID != nil {
		fields = append(fields, zap.String("user_id", *userID))
	}
	
	// Log at different levels based on status code
	switch {
	case statusCode >= 500:
		l.Error("HTTP response", fields...)
	case statusCode >= 400:
		l.Warn("HTTP response", fields...)
	default:
		l.Info("HTTP response", fields...)
	}
}

// LogDatabaseOperation logs a database operation
func (l *Logger) LogDatabaseOperation(operation, table string, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
	}
	
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("Database operation failed", fields...)
	} else {
		l.Debug("Database operation completed", fields...)
	}
}

// LogExternalService logs an external service call
func (l *Logger) LogExternalService(service, operation string, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("service", service),
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	}
	
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("External service call failed", fields...)
	} else {
		l.Debug("External service call completed", fields...)
	}
}

// LogBusinessEvent logs a business event
func (l *Logger) LogBusinessEvent(event string, userID string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("user_id", userID),
	}
	
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	
	l.Info("Business event", fields...)
}

// LogSecurityEvent logs a security-related event
func (l *Logger) LogSecurityEvent(event string, userID *string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
	}
	
	if userID != nil {
		fields = append(fields, zap.String("user_id", *userID))
	}
	
	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}
	
	l.Warn("Security event", fields...)
}
