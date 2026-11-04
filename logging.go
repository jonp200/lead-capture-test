package main

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	maxLogSizeMB  = 10
	maxLogBackups = 5
	maxLogAgeDays = 14
)

func setupLogging(environment, path string) (*slog.Logger, func() error, error) {
	if strings.EqualFold(strings.TrimSpace(environment), "development") {
		return slog.New(slog.NewTextHandler(os.Stderr, nil)), func() error { return nil }, nil
	}
	if path == "" {
		path = filepath.Join("logs", "leads.log")
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() { //nolint:gosec // Operator-configured file.
		return nil, nil, errors.New("could not open log file")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil { //nolint:gosec // Operator-configured directory.
		return nil, nil, errors.New("could not create log directory")
	}
	file := &lumberjack.Logger{
		Filename: path, MaxSize: maxLogSizeMB, MaxBackups: maxLogBackups,
		MaxAge: maxLogAgeDays, Compress: true,
	}
	// Open eagerly so an unwritable destination fails startup, before slog can discard write errors.
	if _, err := file.Write(nil); err != nil {
		_ = file.Close()
		return nil, nil, errors.New("could not open log file")
	}
	return slog.New(slog.NewJSONHandler(file, nil)), file.Close, nil
}

// serverLogWriter keeps net/http diagnostics structured without exposing raw errors.
type serverLogWriter struct {
	logger *slog.Logger
}

func (writer serverLogWriter) Write(data []byte) (int, error) {
	writer.logger.Error("HTTP server diagnostic")
	return len(data), nil
}
