package observability

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger represents a logger
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// StructuredLog represents a structured log entry
type StructuredLog struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
}

// StructuredLogger represents a structured logger
type StructuredLogger struct {
	level LogLevel
	logs  chan StructuredLog
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level LogLevel) *StructuredLogger {
	sl := &StructuredLogger{
		level: level,
		logs:  make(chan StructuredLog, 100),
	}
	go sl.processLogs()
	return sl
}

// Log logs a structured log entry
func (sl *StructuredLogger) Log(level string, message string, fields map[string]interface{}) {
	sl.logs <- StructuredLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}
}

// processLogs processes log entries
func (sl *StructuredLogger) processLogs() {
	for log := range sl.logs {
		fmt.Printf("[%s] %s %v\n", log.Level, log.Message, log.Fields)
	}
}

// Close closes the structured logger
func (sl *StructuredLogger) Close() {
	close(sl.logs)
}

