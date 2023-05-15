package log

import (
	"go.uber.org/zap"
)

// Fields to be added to a logger
type Fields map[string]interface{}

// Logger contains logger and fields
type Logger struct {
	logger *zap.SugaredLogger
	fields []interface{}
}

var zapSugaredLogger *zap.SugaredLogger

func init() {
	zapLogger, _ := zap.NewProduction(zap.AddCallerSkip(1))
	zapSugaredLogger = zapLogger.Sugar()
}

// Log returns an empty field logger
func Log() Logger {
	return Logger{
		logger: zapSugaredLogger,
		fields: []interface{}{},
	}
}

// WithField add a key/value pair to its fields
func (l Logger) WithField(key string, value interface{}) Logger {
	l.fields = append(l.fields, key, value)
	return l
}

// WithField add multiple key/value pairs to its fields
func (l Logger) WithFields(kvs Fields) Logger {
	for k, v := range kvs {
		l = l.WithField(k, v)
	}
	return l
}

// Debug log
func (l Logger) Debug(args ...interface{}) {
	l.logger.With(l.fields...).Debug(args...)
}

// Info log
func (l Logger) Info(args ...interface{}) {
	l.logger.With(l.fields...).Info(args...)
}

// Warn log
func (l Logger) Warn(args ...interface{}) {
	l.logger.With(l.fields...).Warn(args...)
}

// Error log
func (l Logger) Error(args ...interface{}) {
	l.logger.With(l.fields...).Error(args...)
}

// Pance log
func (l Logger) Panic(args ...interface{}) {
	l.logger.With(l.fields...).Panic(args...)
}
