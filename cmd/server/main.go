package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jo-hoe/calendar-manager/internal/api"
	"github.com/jo-hoe/calendar-manager/internal/calendar"
	"github.com/jo-hoe/calendar-manager/internal/calendar/google"
	"github.com/jo-hoe/calendar-manager/internal/config"
	"github.com/jo-hoe/calendar-manager/internal/ical"
)

func main() {
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	level := config.ParseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	provider, err := newCalendarProvider(cfg.Calendar)
	if err != nil {
		log.Fatalf("failed to create calendar provider: %v", err)
	}

	reader := ical.NewReader(cfg.Calendar.ICalURL)

	server := api.NewServer(cfg.Port, provider, reader, logger)

	go func() {
		logger.Info("starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	api.GracefulShutdown(context.Background(), server, logger)
}

func newCalendarProvider(cfg config.CalendarConfig) (calendar.Provider, error) {
	switch cfg.Provider {
	case "google":
		return google.New(cfg.Google)
	default:
		return nil, fmt.Errorf("unsupported calendar provider: %s", cfg.Provider)
	}
}
