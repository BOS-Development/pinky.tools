// Package log provides configurable logging. It will detect if the process is
// running in kubernetes by searching for the "KUBERNETES_SERVICE_HOST"
// environment variable. If it is running in kubernetes it will output logs to
// stdout using json. If it is not running in kubernetes it will output logs in
// a standard single line readable format.
//
// Additionally, you can set a LOG_LEVEL environment value to any of the
// following values, to retrieve only log levels from that level and above. The
// default log level is INFO for running in kubernetes and DEBUG when not.
//
// FATAL
// ERROR
// WARN
// INFO
// DEBUG
package log

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
)

func init() {
	var config zap.Config

	_, inKubernetes := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if inKubernetes {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		switch strings.ToLower(logLevel) {
		case "debug":
			config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		case "info":
			config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		case "warn":
			config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
		case "error":
			config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		case "fatal":
			config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
		}
	}

	l, _ := config.Build()
	logger = l.Sugar()
}

// Debug logs a message with some additional context.
func Debug(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// Info logs a message with some additional context.
func Info(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warn logs a message with some additional context.
func Warn(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

// Error logs a message with some additional context.
func Error(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// Fatal logs a message with some additional context, then calls os.Exit.
func Fatal(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}