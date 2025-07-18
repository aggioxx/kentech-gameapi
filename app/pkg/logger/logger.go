package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

// Logger is a simple wrapper around zap.SugaredLogger to provide structured logging.
// keeping the interface simple for easy use across the application.
// loglevel can be set via the LOG_LEVEL environment variable.
type Logger struct {
	sugar *zap.SugaredLogger
}

func New() *Logger {
	level := zapcore.InfoLevel
	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	logger, _ := cfg.Build()
	return &Logger{
		sugar: logger.Sugar(),
	}
}

func (l *Logger) Info(message string) {
	l.sugar.Info(message)
}

func (l *Logger) Error(message string) {
	l.sugar.Error(message)
}

func (l *Logger) Warn(message string) {
	l.sugar.Warn(message)
}

func (l *Logger) Debug(message string) {
	l.sugar.Debug(message)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.sugar.Warnf(format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.sugar.Debugf(format, args...)
}
