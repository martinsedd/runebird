package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"runebird/internal/config"
	"runebird/internal/email"
	"runebird/internal/logger"
	"runebird/internal/rate"
	"runebird/internal/scheduler"
	"runebird/internal/templates"
)

type Server struct {
	cfg         *config.Config
	logger      *logger.Logger
	sender      *email.Sender
	templates   *templates.TemplateManager
	rateLimiter *rate.Limiter
	scheduler   *scheduler.Scheduler
	httpServer  *http.Server

	emailsSentTotal      *prometheus.CounterVec
	emailsFailedTotal    *prometheus.CounterVec
	emailsScheduledTotal *prometheus.CounterVec
}

type SendRequest struct {
	Template   string                 `json:"template"`
	Recipients []string               `json:"recipients"`
	Data       map[string]interface{} `json:"data"`
}

type ScheduleRequest struct {
	Template   string                 `json:"template"`
	Recipients []string               `json:"recipients"`
	SendAt     time.Time              `json:"send_at"`
	Data       map[string]interface{} `json:"data"`
}

func New(cfg *config.Config, log *logger.Logger, sender *email.Sender, tm *templates.TemplateManager, rl *rate.Limiter, sched *scheduler.Scheduler) *Server {
	emailsSentTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "runebird_emails_sent_total",
			Help: "Total number of emails sent successfully",
		},
		[]string{"template"},
	)
	emailsFailedTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "runebird_emails_failed_total",
			Help: "Total number of emails failed to send",
		},
		[]string{"template"},
	)
	emailsScheduledTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "runebird_emails_scheduled_total",
			Help: "Total number of emails scheduled for future sending",
		},
		[]string{"template"},
	)

	prometheus.MustRegister(emailsSentTotal)
	prometheus.MustRegister(emailsFailedTotal)
	prometheus.MustRegister(emailsScheduledTotal)

	srv := &Server{
		cfg:                  cfg,
		logger:               log,
		sender:               sender,
		templates:            tm,
		rateLimiter:          rl,
		scheduler:            sched,
		emailsSentTotal:      emailsSentTotal,
		emailsFailedTotal:    emailsFailedTotal,
		emailsScheduledTotal: emailsScheduledTotal,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/send", srv.handleSend)
	mux.HandleFunc("/schedule", srv.handleSchedule)
	mux.Handle("/metrics", promhttp.Handler())

	srv.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: mux,
	}

	return srv
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.Int("port", s.cfg.Server.Port))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start HTTP server: %v", err)
	}
	return nil
}

func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down HTTP server")
	if err := s.httpServer.Close(); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %v", err)
	}
	return nil
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Template == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}
	if len(req.Recipients) == 0 {
		http.Error(w, "At least one recipient is required", http.StatusBadRequest)
		return
	}

	body, subject, err := s.templates.Render(req.Template, req.Data)
	if err != nil {
		s.logger.Error("Failed to render template", zap.String("template", req.Template), zap.Error(err))
		s.emailsFailedTotal.WithLabelValues(req.Template).Inc()
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}

	if subject == "" {
		subject = fmt.Sprintf("Email from RuneBird (%s)", req.Template)
	}

	if s.rateLimiter.CanSend() {
		if err := s.sender.Send(req.Recipients, subject, body); err != nil {
			s.logger.Error("Failed to send email", zap.String("template", req.Template), zap.Any("recipients", req.Recipients), zap.Error(err))
			s.emailsFailedTotal.WithLabelValues(req.Template).Inc()
			http.Error(w, fmt.Sprintf("Failed to send email: %v", err), http.StatusInternalServerError)
			return
		}
		if err := s.rateLimiter.ConsumeToken(); err != nil {
			s.logger.Error("Failed to consume rate limiter token", zap.String("template", req.Template), zap.Error(err))
		}
		s.logger.Info("Email sent successfully", zap.String("template", req.Template), zap.Any("recipients", req.Recipients))
		s.emailsSentTotal.WithLabelValues(req.Template).Inc()
	} else {
		s.rateLimiter.QueueEmail(req.Recipients, subject, body)
		s.logger.Info("Email queued due to rate limit", zap.String("template", req.Template), zap.Any("recipients", req.Recipients))
		s.emailsSentTotal.WithLabelValues(req.Template).Inc()
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "success"}`))
}

func (s *Server) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode schedule request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Template == "" {
		http.Error(w, "Template name is required", http.StatusBadRequest)
		return
	}
	if len(req.Recipients) == 0 {
		http.Error(w, "At least one recipient is required", http.StatusBadRequest)
		return
	}
	if req.SendAt.IsZero() {
		http.Error(w, "SendAt time is required", http.StatusBadRequest)
		return
	}
	if req.SendAt.Before(time.Now().UTC()) {
		http.Error(w, "SendAt time must be in the future", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("sched-%d", time.Now().UnixNano())

	if err := s.scheduler.Schedule(id, req.Template, req.Recipients, req.Data, req.SendAt); err != nil {
		s.logger.Error("Failed to schedule email", zap.String("id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to schedule email: %v", err), http.StatusInternalServerError)
		return
	}

	s.emailsScheduledTotal.WithLabelValues(req.Template).Inc()
	s.logger.Info("Email scheduled successfully", zap.String("id", id), zap.Any("recipients", req.Recipients), zap.Time("send_at", req.SendAt))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"status": "success", "task_id": "%s"}`, id)))
}
