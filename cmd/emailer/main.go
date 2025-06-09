package main

import (
	"fmt"
	"go.uber.org/zap"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	"runebird/internal/config"
	"runebird/internal/email"
	"runebird/internal/logger"
	"runebird/internal/rate"
	"runebird/internal/scheduler"
	"runebird/internal/server"
	"runebird/internal/templates"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	log, err := logger.New(&cfg.Logging)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	defer func(log *logger.Logger) {
		err := log.Close()
		if err != nil {
			_, err := fmt.Fprintf(os.Stderr, "Failed to close logger: %v\n", err)
			if err != nil {
				panic(err)
			}
		}
	}(log)

	sender, err := email.New(&cfg.SMTP)
	if err != nil {
		log.Error("Failed to initialize email sender", zap.Error(err))
		os.Exit(1)
	}

	tm, err := templates.New(&cfg.Templates)
	if err != nil {
		log.Error("Failed to initialize template manager", zap.Error(err))
		tm = &templates.TemplateManager{Templates: make(map[string]*template.Template)}
	}

	rl, err := rate.New(&cfg.RateLimit, log)
	if err != nil {
		log.Error("Failed to initialize rate limiter", zap.Error(err))
		os.Exit(1)
	}
	rl.Start()
	defer rl.Stop()

	sched := scheduler.New(log, sender, tm, rl)
	sched.Start()
	defer sched.Stop()

	srv := server.New(cfg, log, sender, tm, rl, sched)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("Failed to start HTTP server", zap.Error(err))
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Received shutdown signal, stopping services")
	if err := srv.Shutdown(); err != nil {
		log.Error("Failed to shutdown HTTP server", zap.Error(err))
	}
}
