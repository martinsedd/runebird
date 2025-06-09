package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"runebird/internal/config"
	"runebird/internal/email"
	"runebird/internal/logger"
	"runebird/internal/rate"
	"runebird/internal/scheduler"
	"runebird/internal/templates"
)

func setupTestServer(t *testing.T) *httptest.Server {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		SMTP: config.SMTPConfig{
			Host:        "smtp.example.com",
			Port:        587,
			Username:    "user",
			Password:    "pass",
			FromAddress: "from@example.com",
		},
		Templates: config.TemplatesConfig{Path: "./test_templates"},
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

	sched := scheduler.New(log, sender, tm, rl)

	srv := New(cfg, log, sender, tm, rl, sched)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/send":
			srv.handleSend(w, r)
		case "/schedule":
			srv.handleSchedule(w, r)
		case "/metrics":
			promhttp.Handler().ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	return testServer
}

func TestServer(t *testing.T) {
	testServer := setupTestServer(t)
	defer testServer.Close()

	t.Run("SendEndpointInvalidMethod", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/send")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got: %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("SendEndpointInvalidJSON", func(t *testing.T) {
		body := bytes.NewBufferString("invalid json")
		resp, err := http.Post(testServer.URL+"/send", "application/json", body)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("SendEndpointMissingFields", func(t *testing.T) {
		req := SendRequest{
			Template:   "",
			Recipients: []string{},
			Data:       map[string]interface{}{},
		}
		body, _ := json.Marshal(req)
		resp, err := http.Post(testServer.URL+"/send", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("SendEndpointTemplateNotFound", func(t *testing.T) {
		req := SendRequest{
			Template:   "nonexistent",
			Recipients: []string{"test@example.com"},
			Data:       map[string]interface{}{"Name": "Alice"},
		}
		body, _ := json.Marshal(req)
		resp, err := http.Post(testServer.URL+"/send", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("ScheduleEndpointInvalidMethod", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/schedule")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got: %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("ScheduleEndpointInvalidJSON", func(t *testing.T) {
		body := bytes.NewBufferString("invalid json")
		resp, err := http.Post(testServer.URL+"/schedule", "application/json", body)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ScheduleEndpointMissingFields", func(t *testing.T) {
		req := ScheduleRequest{
			Template:   "",
			Recipients: []string{},
			SendAt:     time.Time{},
			Data:       map[string]interface{}{},
		}
		body, _ := json.Marshal(req)
		resp, err := http.Post(testServer.URL+"/schedule", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ScheduleEndpointPastSendAt", func(t *testing.T) {
		req := ScheduleRequest{
			Template:   "welcome",
			Recipients: []string{"test@example.com"},
			SendAt:     time.Now().UTC().Add(-time.Hour),
			Data:       map[string]interface{}{"Name": "Alice"},
		}
		body, _ := json.Marshal(req)
		resp, err := http.Post(testServer.URL+"/schedule", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("ScheduleEndpointSuccess", func(t *testing.T) {
		req := ScheduleRequest{
			Template:   "welcome",
			Recipients: []string{"test@example.com"},
			SendAt:     time.Now().UTC().Add(time.Hour), // Future time
			Data:       map[string]interface{}{"Name": "Alice"},
		}
		body, _ := json.Marshal(req)
		resp, err := http.Post(testServer.URL+"/schedule", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got: %d", http.StatusOK, resp.StatusCode)
		}

		var response map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			if status, ok := response["status"]; !ok || status != "success" {
				t.Errorf("expected response status 'success', got: %v", response)
			}
			if _, ok := response["task_id"]; !ok {
				t.Error("expected response to contain task_id, but it was not found")
			}
		}
	})

	t.Run("MetricsEndpoint", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/metrics")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Printf("failed to close response body: %v", err)
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got: %d", http.StatusOK, resp.StatusCode)
		}
	})
}
