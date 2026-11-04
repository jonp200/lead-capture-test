package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProductionLogging(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "leads.log")
	for _, environment := range []string{"production", "", "staging"} {
		logger, closeLog, err := setupLogging(environment, path)
		if err != nil {
			t.Fatal(err)
		}
		logger.Info("test event", "imported", 3)
		if err = closeLog(); err != nil {
			t.Fatal(err)
		}
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	for range 3 {
		var event map[string]any
		if err = decoder.Decode(&event); err != nil {
			t.Fatal("logs must append valid JSON records:", err)
		}
		if event["msg"] != "test event" || event["level"] != "INFO" ||
			event["imported"] != float64(3) || event["time"] == nil {
			t.Fatalf("unexpected structured log: %v", event)
		}
	}
}

func TestLoggingSetupFailure(t *testing.T) {
	directory := t.TempDir()
	if _, _, err := setupLogging("production", directory); err == nil {
		t.Fatal("production must fail if the log file cannot be opened")
	}
}

func TestDevelopmentLogging(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unused.log")
	_, closeLog, err := setupLogging("development", path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeLog()
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("development logging should not create a file")
	}
}

func TestProductionLogRotation(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "leads.log")
	const maxBytes = maxLogSizeMB * 1024 * 1024
	// Leave room to open the file, but not enough for the next complete JSON record.
	seed := `{"msg":"` + strings.Repeat("x", maxBytes-100) + "\"}\n"
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(directory, "leads-2000-01-01T00-00-00.000.log.gz")
	if err := os.WriteFile(stale, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	logger, closeLog, err := setupLogging("", path)
	if err != nil {
		t.Fatal(err)
	}
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "after rotation", 0)
	record.Add("detail", strings.Repeat("y", 256))
	if err = logger.Handler().Handle(t.Context(), record); err != nil {
		_ = closeLog()
		t.Fatal("write that triggers rotation:", err)
	}
	if err = closeLog(); err != nil {
		t.Fatal(err)
	}
	active, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	if err = json.Unmarshal(active, &event); err != nil || event["msg"] != "after rotation" {
		t.Fatal("new record must be intact in the active log:", err)
	}
	// Compression and retention cleanup run asynchronously.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		archives, globErr := filepath.Glob(filepath.Join(directory, "leads-*.log.gz"))
		if globErr != nil {
			t.Fatal(globErr)
		}
		_, statErr := os.Stat(stale)
		if len(archives) == 1 && os.IsNotExist(statErr) {
			if contents, readErr := readGzip(archives[0]); readErr == nil && string(contents) == seed {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("rotation must preserve the old log in gzip and remove expired backups")
}

func readGzip(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
