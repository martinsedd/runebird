package email

import (
	"testing"

	"runebird/internal/config"
)

func TestSender(t *testing.T) {
	t.Run("NewSenderValidConfig", func(t *testing.T) {
		cfg := &config.SMTPConfig{
			Host:        "smtp.example.com",
			Port:        587,
			Username:    "user",
			Password:    "pass",
			FromAddress: "from@example.com",
		}
		sender, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if sender == nil {
			t.Fatal("expected sender to be initialized, got nil")
		}
	})

	t.Run("NewSenderInvalidConfig", func(t *testing.T) {
		cfg := &config.SMTPConfig{
			Host:        "",
			Port:        0,
			Username:    "",
			Password:    "",
			FromAddress: "",
		}
		_, err := New(cfg)
		if err == nil {
			t.Fatal("expected error for invalid config, got none")
		}
	})

	t.Run("SendNoRecipients", func(t *testing.T) {
		cfg := &config.SMTPConfig{
			Host:        "smtp.example.com",
			Port:        587,
			Username:    "user",
			Password:    "pass",
			FromAddress: "from@example.com",
		}
		sender, err := New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		err = sender.Send([]string{}, "Test Subject", "<p>Test Body</p>")
		if err == nil {
			t.Fatal("expected error for no recipients, got none")
		}
	})

	t.Run("SendEmailMock", func(t *testing.T) {
		t.Skip("Skipping actual SMTP send test; requires mock server setup")
	})
}
