package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	createTempYAML := func(t *testing.T, content string) string {
		t.Helper()
		tmpFile, err := os.CreateTemp("", "emailer-*.yaml")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer func(tmpFile *os.File) {
			err := tmpFile.Close()
			if err != nil {
				fmt.Printf("failed to close temp file: %v", err)
			}
		}(tmpFile)

		_, err = tmpFile.WriteString(content)
		if err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		return tmpFile.Name()
	}

	t.Run("ValidConfigWithDefaults", func(t *testing.T) {
		content := `
server:
  port: 8080
smtp:
  host: "smtp.example.com"
  username: "user"
  password: "password"
  from_address: "test@example.com"
`
		tmpPath := createTempYAML(t, content)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Printf("failed to remove temp file: %v", err)
			}
		}(tmpPath)

		err := os.Setenv("EMAILER_CONFIG_PATH", tmpPath)
		if err != nil {
			t.Fatalf("failed to set env var: %v", err)
		}
		defer func() {
			err := os.Unsetenv("EMAILER_CONFIG_PATH")
			if err != nil {
				fmt.Printf("failed to unset env var: %v", err)
			}
		}()

		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Check defaults
		if cfg.Templates.Path != "./templates" {
			t.Errorf("expected default templates path './templates', got: %s", cfg.Templates.Path)
		}
		if cfg.RateLimit.PerHour != 100 {
			t.Errorf("expected default rate limit per hour 100, got: %d", cfg.RateLimit.PerHour)
		}
		if cfg.RateLimit.Burst != 5 {
			t.Errorf("expected default burst 5, got: %d", cfg.RateLimit.Burst)
		}
		if cfg.Logging.Level != "info" {
			t.Errorf("expected default log level 'info', got: %s", cfg.Logging.Level)
		}
	})

	t.Run("InvalidPort", func(t *testing.T) {
		content := `
server:
  port: 70000
smtp:
  host: "smtp.example.com"
  port: 587
  username: "user"
  password: "pass"
  from_address: "test@example.com"
`
		tmpPath := createTempYAML(t, content)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Printf("failed to remove temp file: %v", err)
			}
		}(tmpPath)

		err := os.Setenv("EMAILER_CONFIG_PATH", tmpPath)
		if err != nil {
			t.Fatalf("failed to set env var: %v", err)
		}
		defer func() {
			err := os.Unsetenv("EMAILER_CONFIG_PATH")
			if err != nil {
				fmt.Printf("failed to unset env var: %v", err)
			}
		}()

		_, err = Load()
		if err == nil {
			t.Fatal("expected error for invalid port, got none")
		}
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		content := `
server:
  port: 8080
smtp:
  port: 587
`
		tmpPath := createTempYAML(t, content)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Printf("failed to remove temp file: %v", err)
			}
		}(tmpPath)

		err := os.Setenv("EMAILER_CONFIG_PATH", tmpPath)
		if err != nil {
			t.Fatalf("failed to set env var: %v", err)
		}
		defer func() {
			err := os.Unsetenv("EMAILER_CONFIG_PATH")
			if err != nil {
				fmt.Printf("failed to unset env var: %v", err)
			}
		}()

		_, err = Load()
		if err == nil {
			t.Fatal("expected error for missing required SMTP fields, got none")
		}
	})

	t.Run("InvalidLogLevel", func(t *testing.T) {
		content := `
server:
  port: 8080
smtp:
  host: "smtp.example.com"
  port: 587
  username: "user"
  password: "pass"
  from_address: "test@example.com"
logging:
  level: "invalid"
`
		tmpPath := createTempYAML(t, content)
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Printf("failed to remove temp file: %v", err)
			}
		}(tmpPath)

		err := os.Setenv("EMAILER_CONFIG_PATH", tmpPath)
		if err != nil {
			t.Fatalf("failed to set env var: %v", err)
		}
		defer func() {
			err := os.Unsetenv("EMAILER_CONFIG_PATH")
			if err != nil {
				fmt.Printf("failed to unset env var: %v", err)
			}
		}()

		_, err = Load()
		if err == nil {
			t.Fatal("expected error for invalid log level, got none")
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		err := os.Setenv("EMAILER_CONFIG_PATH", "nonexistent.yaml")
		if err != nil {
			t.Fatalf("failed to set env var: %v", err)
		}
		defer func() {
			err := os.Unsetenv("EMAILER_CONFIG_PATH")
			if err != nil {
				fmt.Printf("failed to unset env var: %v", err)
			}
		}()

		_, err = Load()
		if err == nil {
			t.Fatal("expected error for nonexistent file, got none")
		}
	})

	t.Run("DefaultPath", func(t *testing.T) {
		// Ensure there's an emailer.yaml in the test directory or mock it
		content := `
server:
  port: 8080
smtp:
  host: "smtp.example.com"
  port: 587
  username: "user"
  password: "pass"
  from_address: "test@example.com"
`
		tmpPath := filepath.Join(".", "emailer.yaml")
		err := os.WriteFile(tmpPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create default emailer.yaml: %v", err)
		}
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				fmt.Printf("failed to remove temp file: %v", err)
			}
		}(tmpPath)

		err = os.Unsetenv("EMAILER_CONFIG_PATH")
		if err != nil {
			t.Fatalf("failed to unset env var: %v", err)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error with default path, got: %v", err)
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("expected server port 8080, got: %d", cfg.Server.Port)
		}
	})
}
