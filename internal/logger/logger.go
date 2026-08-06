package logger

import (
	"log"
	"os"
)

// Logger defines logging interface
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
}

// StdoutLogger implements Logger dengan simple stdout output
type StdoutLogger struct {
	logger *log.Logger
}

// NewStdoutLogger creates logger yang writes ke stdout
func NewStdoutLogger() *StdoutLogger {
	return &StdoutLogger{logger: log.New(os.Stdout, "", log.LstdFlags)}
}

// Info logs informational message
func (l *StdoutLogger) Info(msg string, fields map[string]interface{}) {
	l.logger.Printf("[INFO] %s | %v\n", msg, fields)
}

// Warn logs warning message
func (l *StdoutLogger) Warn(msg string, fields map[string]interface{}) {
	l.logger.Printf("[WARN] %s | %v\n", msg, fields)
}

// Error logs error message
func (l *StdoutLogger) Error(msg string, fields map[string]interface{}) {
	l.logger.Printf("[ERROR] %s | %v\n", msg, fields)
}