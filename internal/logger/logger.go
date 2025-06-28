package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	level   LogLevel
	logFile *os.File
	logger  *log.Logger
	context map[string]interface{}
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) (*Logger, error) {
	// Create log directory
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable is not set")
	}

	logDir := filepath.Join(homeDir, ".run", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("run-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New(file, "", log.LstdFlags)

	return &Logger{
		level:   level,
		logFile: file,
		logger:  logger,
		context: make(map[string]interface{}),
	}, nil
}

// WithContext adds context to the logger
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:   l.level,
		logFile: l.logFile,
		logger:  l.logger,
		context: make(map[string]interface{}),
	}

	// Copy existing context
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	// Add new context
	newLogger.context[key] = value
	return newLogger
}

// WithPackage adds package context to the logger
func (l *Logger) WithPackage(packageName string) *Logger {
	return l.WithContext("package", packageName)
}

// WithOperation adds operation context to the logger
func (l *Logger) WithOperation(operation string) *Logger {
	return l.WithContext("operation", operation)
}

// log formats and writes a log message
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Format message with context
	formattedMessage := fmt.Sprintf(message, args...)

	// Add context to message
	if len(l.context) > 0 {
		contextStr := ""
		for key, value := range l.context {
			if contextStr != "" {
				contextStr += " "
			}
			contextStr += fmt.Sprintf("%s=%v", key, value)
		}
		formattedMessage = fmt.Sprintf("[%s] %s", contextStr, formattedMessage)
	}

	// Write to log file
	l.logger.Printf("[%s] %s", level.String(), formattedMessage)

	// Also write to console for ERROR and FATAL levels
	if level >= ERROR {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", level.String(), formattedMessage)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log(FATAL, message, args...)
	os.Exit(1)
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Global logger instance with thread-safe initialization
var (
	globalLogger *Logger
	loggerOnce   sync.Once
)

// InitLogger initializes the global logger
func InitLogger(level LogLevel) error {
	var err error
	globalLogger, err = NewLogger(level)
	return err
}

// GetLogger returns the global logger with thread-safe initialization
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		if globalLogger == nil {
			// Initialize with INFO level if not initialized
			if err := InitLogger(INFO); err != nil {
				// If initialization fails, create a minimal logger
				globalLogger = &Logger{
					level:   INFO,
					context: make(map[string]interface{}),
				}
			}
		}
	})
	return globalLogger
}

// Convenience functions for global logger
func Debug(message string, args ...interface{}) {
	GetLogger().Debug(message, args...)
}

func Info(message string, args ...interface{}) {
	GetLogger().Info(message, args...)
}

func Warn(message string, args ...interface{}) {
	GetLogger().Warn(message, args...)
}

func Error(message string, args ...interface{}) {
	GetLogger().Error(message, args...)
}

func Fatal(message string, args ...interface{}) {
	GetLogger().Fatal(message, args...)
}
