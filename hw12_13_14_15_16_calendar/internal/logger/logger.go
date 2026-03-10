package logger

import (
	"fmt"
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
	file   *os.File
}

func New(level string) (*Logger, error) {
	var slevel slog.Level

	err := slevel.UnmarshalText([]byte(level))
	if err != nil {
		return nil, fmt.Errorf("error parsing level %s: %w", level, err)
	}

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	handler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: &slevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return &Logger{
		logger: logger,
		file:   file,
	}, nil
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
