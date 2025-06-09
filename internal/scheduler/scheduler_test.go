package scheduler

import (
	"testing"
	"time"

	"runebird/internal/config"
	"runebird/internal/email"
	"runebird/internal/logger"
	"runebird/internal/rate"
	"runebird/internal/templates"
)

func setupTestScheduler(t *testing.T) (*Scheduler, *email.Sender, *templates.TemplateManager, *rate.Limiter) {
	cfg := &config.Config{
		SMTP: config.SMTPConfig{
			Host:        "smtp.example.com",
			Port:        587,
			Username:    "user",
			Password:    "pass",
			FromAddress: "from@example.com",
		},
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

	sender, err := email.New(&cfg.SMTP)
	if err != nil {
		t.Fatalf("failed to create email sender: %v", err)
	}

	tm := &templates.TemplateManager{}

	rl, err := rate.New(&cfg.RateLimit, log)
	if err != nil {
		t.Fatalf("failed to create rate limiter: %v", err)
	}

	scheduler := New(log, sender, tm, rl)

	return scheduler, sender, tm, rl
}

func TestScheduler(t *testing.T) {
	t.Run("NewScheduler", func(t *testing.T) {
		scheduler, _, _, _ := setupTestScheduler(t)
		if scheduler == nil {
			t.Fatal("expected scheduler to be initialized, got nil")
		}
	})

	t.Run("StartAndStop", func(t *testing.T) {
		scheduler, _, _, _ := setupTestScheduler(t)
		scheduler.Start()
		time.Sleep(100 * time.Millisecond)
		if !scheduler.isRunning {
			t.Error("expected scheduler to be running after Start")
		}

		scheduler.Stop()
		time.Sleep(100 * time.Millisecond)
		if scheduler.isRunning {
			t.Error("expected scheduler to be stopped after Stop")
		}
	})

	t.Run("ScheduleTask", func(t *testing.T) {
		scheduler, _, _, _ := setupTestScheduler(t)
		id := "test-task-1"
		template := "welcome"
		recipients := []string{"test@example.com"}
		data := map[string]interface{}{"Name": "Alice"}
		sendAt := time.Now().UTC().Add(time.Minute * 5)

		err := scheduler.Schedule(id, template, recipients, data, sendAt)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		scheduler.mu.Lock()
		if _, exists := scheduler.tasks[id]; !exists {
			t.Error("expected task to be scheduled, but it was not found")
		}
		if task, exists := scheduler.tasks[id]; exists {
			if task.SendAt != sendAt.UTC() {
				t.Errorf("expected SendAt to be %v, got: %v", sendAt, task.SendAt)
			}
		}
		scheduler.mu.Unlock()
	})

	t.Run("ScheduleDuplicateTask", func(t *testing.T) {
		scheduler, _, _, _ := setupTestScheduler(t)
		id := "test-task-2"
		template := "welcome"
		recipients := []string{"test@example.com"}
		data := map[string]interface{}{"Name": "Alice"}
		sendAt := time.Now().UTC().Add(time.Minute * 5)

		err := scheduler.Schedule(id, template, recipients, data, sendAt)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		err = scheduler.Schedule(id, template, recipients, data, sendAt)
		if err == nil {
			t.Fatal("expected error for duplicate task ID, got none")
		}
	})

	t.Run("ProcessTask", func(t *testing.T) {
		scheduler, _, _, _ := setupTestScheduler(t)
		id := "test-task-3"
		template := "nonexistent"
		recipients := []string{"test@example.com"}
		data := map[string]interface{}{"Name": "Alice"}
		sendAt := time.Now().UTC().Add(-time.Minute)

		err := scheduler.Schedule(id, template, recipients, data, sendAt)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		scheduler.mu.Lock()
		task, exists := scheduler.tasks[id]
		scheduler.mu.Unlock()
		if exists {
			scheduler.processTask(id, task)
		}
	})
}
