package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

const (
	LoggerErrorLevel = "error"
	LoggerWarnLevel  = "warn"
	LoggerInfoLevel  = "info"
	LoggerDebugLevel = "debug"
)

type ILogger interface {
	Infof(message string, args ...interface{})
	Warnf(message string, args ...interface{})
	Errorf(message string, args ...interface{})
	Fatalf(message string, args ...interface{})
	Debugf(message string, args ...interface{})
}

type Logger struct {
	logger *zerolog.Logger
}

func NewLogger(logLevel string, w io.Writer) ILogger {
	var l zerolog.Level
	switch logLevel {
	case LoggerErrorLevel:
		l = zerolog.ErrorLevel
	case LoggerWarnLevel:
		l = zerolog.WarnLevel
	case LoggerInfoLevel:
		l = zerolog.InfoLevel
	case LoggerDebugLevel:
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(w).With().
		Timestamp().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).
		Logger()
	return &Logger{
		logger: &logger,
	}
}

func (l *Logger) Infof(message string, args ...interface{}) {
	l.logger.Info().Msgf(message, args...)
}

func (l *Logger) Warnf(message string, args ...interface{}) {
	l.logger.Warn().Msgf(message, args...)
}

func (l *Logger) Errorf(message string, args ...interface{}) {
	l.logger.Error().Msgf(message, args...)
}

func (l *Logger) Fatalf(message string, args ...interface{}) {
	l.logger.Fatal().Msgf(message, args...)
	os.Exit(1)
}

func (l *Logger) Debugf(message string, args ...interface{}) {
	l.logger.Debug().Msgf(message, args...)
}
