package logging

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "development logger",
			config: Config{
				Level:      "debug",
				Format:     "console",
				OutputPath: "stdout",
			},
			want: "debug",
		},
		{
			name: "production logger",
			config: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "stdout",
			},
			want: "info",
		},
		{
			name: "default logger",
			config: Config{
				Level:      "invalid",
				Format:     "",
				OutputPath: "",
			},
			want: "info", // Should default to info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if logger == nil {
				t.Fatal("Expected logger to be created, got nil")
			}
		})
	}
}

func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment()
	if err != nil {
		t.Fatalf("NewDevelopment() error = %v", err)
	}
	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestNewProduction(t *testing.T) {
	logger, err := NewProduction()
	if err != nil {
		t.Fatalf("NewProduction() error = %v", err)
	}
	if logger == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestLoggerWithContext(t *testing.T) {
	// Create a logger with observer for testing
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	// Test WithRequestID
	requestLogger := logger.WithRequestID("req-123")
	requestLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	// Check if request_id field is present
	found := false
	for _, field := range logs[0].Context {
		if field.Key == "request_id" && field.String == "req-123" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected request_id field in log context")
	}
}

func TestLoggerWithUserID(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	userLogger := logger.WithUserID("user-456")
	userLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	found := false
	for _, field := range logs[0].Context {
		if field.Key == "user_id" && field.String == "user-456" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected user_id field in log context")
	}
}

func TestLoggerWithOperation(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	opLogger := logger.WithOperation("test-operation")
	opLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	found := false
	for _, field := range logs[0].Context {
		if field.Key == "operation" && field.String == "test-operation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected operation field in log context")
	}
}

func TestLoggerWithDuration(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	duration := 100 * time.Millisecond
	durationLogger := logger.WithDuration(duration)
	durationLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	found := false
	for _, field := range logs[0].Context {
		if field.Key == "duration" && field.Type == zapcore.DurationType {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected duration field in log context")
	}
}

func TestLoggerWithError(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	testErr := errors.New("test error")
	errorLogger := logger.WithError(testErr)
	errorLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	found := false
	for _, field := range logs[0].Context {
		if field.Key == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error field in log context")
	}
}

func TestLoggerWithField(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	fieldLogger := logger.WithField("custom_field", "custom_value")
	fieldLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	found := false
	for _, field := range logs[0].Context {
		if field.Key == "custom_field" && field.String == "custom_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected custom_field in log context")
	}
}

func TestLoggerWithFields(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}
	
	fieldsLogger := logger.WithFields(fields)
	fieldsLogger.Info("test message")
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	// Check that all fields are present
	expectedFields := map[string]bool{
		"field1": false,
		"field2": false,
		"field3": false,
	}
	
	for _, field := range logs[0].Context {
		if _, exists := expectedFields[field.Key]; exists {
			expectedFields[field.Key] = true
		}
	}
	
	for field, found := range expectedFields {
		if !found {
			t.Errorf("Expected field %s in log context", field)
		}
	}
}

func TestLogRequest(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	userID := "user-123"
	logger.LogRequest("GET", "/api/users", "Mozilla/5.0", "127.0.0.1", &userID)
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	log := logs[0]
	if log.Message != "HTTP request received" {
		t.Errorf("Expected message 'HTTP request received', got '%s'", log.Message)
	}
	
	// Check required fields
	requiredFields := map[string]bool{
		"method":      false,
		"path":        false,
		"user_agent":  false,
		"remote_addr": false,
		"user_id":     false,
	}
	
	for _, field := range log.Context {
		if _, exists := requiredFields[field.Key]; exists {
			requiredFields[field.Key] = true
		}
	}
	
	for field, found := range requiredFields {
		if !found {
			t.Errorf("Expected field %s in log context", field)
		}
	}
}

func TestLogResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expectedLevel zapcore.Level
	}{
		{
			name:       "success response",
			statusCode: 200,
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name:       "client error response",
			statusCode: 400,
			expectedLevel: zapcore.WarnLevel,
		},
		{
			name:       "server error response",
			statusCode: 500,
			expectedLevel: zapcore.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, recorded := observer.New(zapcore.DebugLevel)
			logger := &Logger{Logger: zap.New(core)}

			userID := "user-123"
			logger.LogResponse("GET", "/api/users", tt.statusCode, 100*time.Millisecond, &userID)
			
			logs := recorded.All()
			if len(logs) == 0 {
				t.Fatal("Expected log entry, got none")
			}
			
			log := logs[0]
			if log.Level != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, log.Level)
			}
			
			if log.Message != "HTTP response" {
				t.Errorf("Expected message 'HTTP response', got '%s'", log.Message)
			}
		})
	}
}

func TestLogDatabaseOperation(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := &Logger{Logger: zap.New(core)}

	// Test successful operation
	logger.LogDatabaseOperation("SELECT", "users", 50*time.Millisecond, nil)
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	log := logs[0]
	if log.Level != zapcore.DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", log.Level)
	}
	
	if log.Message != "Database operation completed" {
		t.Errorf("Expected message 'Database operation completed', got '%s'", log.Message)
	}
	
	// Test failed operation
	testErr := errors.New("database error")
	logger.LogDatabaseOperation("INSERT", "users", 100*time.Millisecond, testErr)
	
	logs = recorded.All()
	if len(logs) < 2 {
		t.Fatal("Expected at least 2 log entries")
	}
	
	log = logs[1]
	if log.Level != zapcore.ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", log.Level)
	}
	
	if log.Message != "Database operation failed" {
		t.Errorf("Expected message 'Database operation failed', got '%s'", log.Message)
	}
}

func TestLogExternalService(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := &Logger{Logger: zap.New(core)}

	// Test successful service call
	logger.LogExternalService("S3", "UploadImage", 200*time.Millisecond, nil)
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	log := logs[0]
	if log.Level != zapcore.DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", log.Level)
	}
	
	if log.Message != "External service call completed" {
		t.Errorf("Expected message 'External service call completed', got '%s'", log.Message)
	}
	
	// Test failed service call
	testErr := errors.New("service error")
	logger.LogExternalService("Redis", "Set", 50*time.Millisecond, testErr)
	
	logs = recorded.All()
	if len(logs) < 2 {
		t.Fatal("Expected at least 2 log entries")
	}
	
	log = logs[1]
	if log.Level != zapcore.ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", log.Level)
	}
	
	if log.Message != "External service call failed" {
		t.Errorf("Expected message 'External service call failed', got '%s'", log.Message)
	}
}

func TestLogBusinessEvent(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := &Logger{Logger: zap.New(core)}

	details := map[string]interface{}{
		"action": "user_registration",
		"source": "web",
	}
	
	logger.LogBusinessEvent("user_registered", "user-123", details)
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	log := logs[0]
	if log.Level != zapcore.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", log.Level)
	}
	
	if log.Message != "Business event" {
		t.Errorf("Expected message 'Business event', got '%s'", log.Message)
	}
	
	// Check required fields
	requiredFields := map[string]bool{
		"event":   false,
		"user_id": false,
		"action":  false,
		"source":  false,
	}
	
	for _, field := range log.Context {
		if _, exists := requiredFields[field.Key]; exists {
			requiredFields[field.Key] = true
		}
	}
	
	for field, found := range requiredFields {
		if !found {
			t.Errorf("Expected field %s in log context", field)
		}
	}
}

func TestLogSecurityEvent(t *testing.T) {
	core, recorded := observer.New(zapcore.WarnLevel)
	logger := &Logger{Logger: zap.New(core)}

	userID := "user-123"
	details := map[string]interface{}{
		"ip_address": "192.168.1.1",
		"user_agent":  "Mozilla/5.0",
	}
	
	logger.LogSecurityEvent("failed_login", &userID, details)
	
	logs := recorded.All()
	if len(logs) == 0 {
		t.Fatal("Expected log entry, got none")
	}
	
	log := logs[0]
	if log.Level != zapcore.WarnLevel {
		t.Errorf("Expected WarnLevel, got %v", log.Level)
	}
	
	if log.Message != "Security event" {
		t.Errorf("Expected message 'Security event', got '%s'", log.Message)
	}
	
	// Check required fields
	requiredFields := map[string]bool{
		"event":      false,
		"user_id":    false,
		"ip_address": false,
		"user_agent": false,
	}
	
	for _, field := range log.Context {
		if _, exists := requiredFields[field.Key]; exists {
			requiredFields[field.Key] = true
		}
	}
	
	for field, found := range requiredFields {
		if !found {
			t.Errorf("Expected field %s in log context", field)
		}
	}
}
