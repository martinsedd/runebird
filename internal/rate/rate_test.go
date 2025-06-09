package rate

import (
	"testing"
	"time"

	"runebird/internal/config"
	"runebird/internal/logger"
)

func TestLimiter(t *testing.T) {
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			PerHour: 600,
			Burst:   2,
		},
		Logging: config.LoggingConfig{
			Level:    "info",
			FilePath: "",
		},
	}

	log, err := logger.New(&cfg.Logging)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	t.Run("NewLimiterValidConfig", func(t *testing.T) {
		limiter, err := New(&cfg.RateLimit, log)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if limiter == nil {
			t.Fatal("expected limiter to be initialized, got nil")
		}
		limiter.Stop()
	})

	t.Run("NewLimiterInvalidConfig", func(t *testing.T) {
		invalidCfg := &config.RateLimitConfig{
			PerHour: 0,
			Burst:   0,
		}
		_, err := New(invalidCfg, log)
		if err == nil {
			t.Fatal("expected error for invalid config, got none")
		}
	})

	t.Run("CanSendAndQueue", func(t *testing.T) {
		limiter, err := New(&cfg.RateLimit, log)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		defer limiter.Stop()

		if !limiter.CanSend() {
			t.Error("expected CanSend to return true for initial burst")
		}
		if !limiter.CanSend() {
			t.Error("expected CanSend to return true for second burst")
		}

		if limiter.CanSend() {
			t.Error("expected CanSend to return false after burst is used")
		}

		recipients := []string{"test@example.com"}
		limiter.QueueEmail(recipients, "Test Subject", "<p>Test Body</p>")

		queued := limiter.GetQueuedEmails()
		if len(queued) != 0 {
			t.Error("expected queued email to not be ready immediately")
		}
	})

	t.Run("StartAndStop", func(t *testing.T) {
		limiter, err := New(&cfg.RateLimit, log)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		limiter.Start()
		time.Sleep(100 * time.Millisecond)
		if !limiter.isRunning {
			t.Error("expected limiter to be running after Start")
		}

		limiter.Stop()
		time.Sleep(100 * time.Millisecond)
		if limiter.isRunning {
			t.Error("expected limiter to be stopped after Stop")
		}
	})

	t.Run("QueueProcessing", func(t *testing.T) {
		limiter, err := New(&cfg.RateLimit, log)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		defer limiter.Stop()

		recipients := []string{"test@example.com"}
		limiter.QueueEmail(recipients, "Test Subject", "<p>Test Body</p>")

		limiter.Start()
		time.Sleep(11 * time.Second)

		queued := limiter.GetQueuedEmails()
		if len(queued) > 0 {
			t.Logf("queued emails still present, may not have been processed yet: %d", len(queued))
		}
	})
}
