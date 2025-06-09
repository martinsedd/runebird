// Package rate implements global rate limiting for email sending in the RuneBird emailer service.
package rate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"runebird/internal/config"
	"runebird/internal/logger"
)

// EmailTask represents a delayed email sending task.
type EmailTask struct {
	Recipients []string
	Subject    string
	Body       string
	RetryAt    time.Time
}

// Limiter manages rate limiting for email sending with delayed retries.
type Limiter struct {
	limiter   *rate.Limiter
	queue     []EmailTask
	mu        sync.Mutex
	logger    *logger.Logger
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// New creates a new Limiter instance based on the provided rate limit configuration.
func New(cfg *config.RateLimitConfig, log *logger.Logger) (*Limiter, error) {
	if cfg.PerHour <= 0 || cfg.Burst <= 0 {
		return nil, fmt.Errorf("invalid rate limit configuration: per_hour=%d, burst=%d", cfg.PerHour, cfg.Burst)
	}

	// Calculate rate per second from per hour limit
	ratePerSecond := float64(cfg.PerHour) / 3600.0
	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), cfg.Burst)

	ctx, cancel := context.WithCancel(context.Background())

	return &Limiter{
		limiter:   limiter,
		queue:     make([]EmailTask, 0),
		logger:    log,
		isRunning: false,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start begins processing the delayed email queue in a non-blocking manner.
func (l *Limiter) Start() {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return
	}
	l.isRunning = true
	l.mu.Unlock()

	go l.processQueue()
	l.logger.Info("Rate limiter queue processing started")
}

// Stop halts the processing of the delayed email queue.
func (l *Limiter) Stop() {
	l.mu.Lock()
	if !l.isRunning {
		l.mu.Unlock()
		return
	}
	l.isRunning = false
	l.mu.Unlock()

	l.cancel()
	l.logger.Info("Rate limiter queue processing stopped")
}

// CanSend checks if an email can be sent immediately based on the rate limit.
// Returns true if a token is available now without waiting, false if it should be queued.
func (l *Limiter) CanSend() bool {
	reservation := l.limiter.ReserveN(time.Now(), 1)
	if reservation.OK() {
		// If reservation is OK and delay is zero or negative, a token is available now
		if reservation.Delay() <= 0 {
			return true
		}
		// Cancel the reservation since we won't use it (we're not waiting)
		reservation.Cancel()
	}
	return false
}

// ConsumeToken consumes a token from the rate limiter, blocking if necessary until one is available.
func (l *Limiter) ConsumeToken() error {
	return l.limiter.WaitN(l.ctx, 1)
}

// QueueEmail adds an email task to the delayed queue if the rate limit is exceeded.
func (l *Limiter) QueueEmail(recipients []string, subject, body string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	task := EmailTask{
		Recipients: recipients,
		Subject:    subject,
		Body:       body,
		RetryAt:    time.Now().Add(time.Second * 10), // Retry after a short delay
	}
	l.queue = append(l.queue, task)
	l.logger.Info("Email queued due to rate limit", zap.Any("recipients", recipients))
}

// GetQueuedEmails retrieves emails from the queue that are ready to be sent.
// Returns a slice of tasks ready for retry.
func (l *Limiter) GetQueuedEmails() []EmailTask {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	var ready []EmailTask
	var remaining []EmailTask

	for _, task := range l.queue {
		if now.After(task.RetryAt) {
			ready = append(ready, task)
		} else {
			remaining = append(remaining, task)
		}
	}

	l.queue = remaining
	return ready
}

// processQueue runs a background loop to process queued emails when the rate limit allows.
func (l *Limiter) processQueue() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.mu.Lock()
			if !l.isRunning {
				l.mu.Unlock()
				return
			}
			l.mu.Unlock()

			readyTasks := l.GetQueuedEmails()
			for _, task := range readyTasks {
				if l.CanSend() {
					// Here, in a real integration, we would trigger sending the email.
					// For now, log the attempt (integration will be handled in server/email packages).
					l.logger.Info("Processing queued email", zap.Any("recipients", task.Recipients))
					// Reserve a token for sending (in real usage, this would be tied to actual send).
					_ = l.limiter.WaitN(l.ctx, 1)
				} else {
					// Re-queue if still rate-limited
					l.QueueEmail(task.Recipients, task.Subject, task.Body)
				}
			}
		}
	}
}
