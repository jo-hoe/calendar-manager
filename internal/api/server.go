package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jo-hoe/calendar-manager/internal/calendar"
	"github.com/jo-hoe/calendar-manager/internal/ical"
)

func NewServer(port int, provider calendar.Provider, reader *ical.Reader, logger *slog.Logger) *http.Server {
	handler := NewHandler(provider, reader, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /probe", handler.HandleProbe)
	mux.HandleFunc("GET /api/events", handler.HandleListEvents)
	mux.HandleFunc("POST /api/events", handler.HandleCreateEvent)
	mux.HandleFunc("DELETE /api/events/{id}", handler.HandleDeleteEvent)

	var h http.Handler = mux
	h = ContentTypeMiddleware(h)
	h = RecoveryMiddleware(logger)(h)
	h = LoggingMiddleware(logger, "/probe")(h)

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
