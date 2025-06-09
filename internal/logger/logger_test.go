package logger

import (
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"
	"runebird/internal/config"
)

func TestNewLogger(t *testing.T) {
	t.Run("ConsoleOnly", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:    "info",
			FilePath: "",
		}
		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if err := logger.Close(); err != nil {
			t.Logf("ignoring close error: %v", err)
		}

		if logger == nil {
			t.Fatal("expected logger to be initialized, got nil")
		}
	})

	t.Run("ConsoleAndFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")
		cfg := &config.LoggingConfig{
			Level:    "debug",
			FilePath: logFile,
		}
		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if err := logger.Close(); err != nil {
			t.Logf("ignoring close error: %v", err)
		}

		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Errorf("expected log file to be created at %s, but it wasn't", logFile)
		}
	})

	t.Run("InvalidLogLevel", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:    "invalid",
			FilePath: "",
		}
		_, err := New(cfg)
		if err == nil {
			t.Fatal("expected error for invalid log level, got none")
		}
	})

	t.Run("InvalidFilePath", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:    "info",
			FilePath: "/invalid/path/to/log/file.log",
		}
		_, err := New(cfg)
		if err == nil {
			t.Fatal("expected error for invalid file path, got none")
		}
	})

	t.Run("ZapTestLogger", func(t *testing.T) {
		testLogger := zaptest.NewLogger(t)
		defer func(testLogger *zap.Logger) {
			err := testLogger.Sync()
			if err != nil {
				t.Fatalf("failed to sync logger: %v", err)
			}
		}(testLogger)

		if testLogger == nil {
			t.Fatal("expected test logger to be initialized, got nil")
		}
	})
}
