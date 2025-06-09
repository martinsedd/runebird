package scheduler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"runebird/internal/email"
	"runebird/internal/logger"
	"runebird/internal/rate"
	"runebird/internal/templates"
	"sync"
	"time"
)

type ScheduledTask struct {
	ID         string
	Template   string
	Recipients []string
	Data       map[string]interface{}
	SendAt     time.Time
}

type Scheduler struct {
	tasks       map[string]ScheduledTask
	mu          sync.Mutex
	logger      *logger.Logger
	sender      *email.Sender
	templates   *templates.TemplateManager
	rateLimiter *rate.Limiter
	isRunning   bool
	ctx         context.Context
	cancel      context.CancelFunc
}

func New(log *logger.Logger, sender *email.Sender, templates *templates.TemplateManager, rateLimiter *rate.Limiter) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		tasks:       make(map[string]ScheduledTask),
		logger:      log,
		sender:      sender,
		templates:   templates,
		rateLimiter: rateLimiter,
		isRunning:   false,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	go s.processTasks()
	s.logger.Info("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	s.cancel()
	s.logger.Info("Scheduler stopped")
}

func (s *Scheduler) Schedule(id, template string, recipients []string, data map[string]interface{}, sendAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; exists {
		return fmt.Errorf("task with ID %s already exists", id)
	}

	sendAt = sendAt.UTC()

	task := ScheduledTask{
		ID:         id,
		Template:   template,
		Recipients: recipients,
		Data:       data,
		SendAt:     sendAt,
	}

	s.tasks[id] = task
	s.logger.Info("Scheduled email task", zap.String("id", id), zap.Time("send_at", sendAt))
	return nil
}

func (s *Scheduler) processTasks() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			if !s.isRunning {
				s.mu.Unlock()
				return
			}

			now := time.Now().UTC()
			var toDelete []string
			for id, task := range s.tasks {
				if now.After(task.SendAt) || now.Equal(task.SendAt) {
					s.mu.Unlock()
					s.processTask(id, task)
					s.mu.Lock()
					toDelete = append(toDelete, id)
				}
			}
			for _, id := range toDelete {
				delete(s.tasks, id)
			}
			s.mu.Unlock()
		}
	}
}

func (s *Scheduler) processTask(id string, task ScheduledTask) {
	s.logger.Info("Processing scheduled email", zap.String("id", id), zap.Any("recipients", task.Recipients))

	body, subject, err := s.templates.Render(task.Template, task.Data)
	if err != nil {
		s.logger.Error("Failed to render template for scheduled email", zap.String("id", id), zap.Any("recipients", task.Recipients), zap.Error(err))
		return
	}

	if subject == "" {
		subject = fmt.Sprintf("Scheduled email from RuneBird (%s)", task.Template)
	}

	if s.rateLimiter.CanSend() {
		if err := s.sender.Send(task.Recipients, subject, body); err != nil {
			s.logger.Error("Failed to send scheduled email", zap.String("id", id), zap.Any("recipients", task.Recipients), zap.Error(err))
			return
		}
	} else {
		s.rateLimiter.QueueEmail(task.Recipients, subject, body)
		s.logger.Info("Scheduled email queued due to rate limit", zap.String("id", id), zap.String("subject", subject))
	}

}
