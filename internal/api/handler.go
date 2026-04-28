package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jo-hoe/calendar-manager/internal/calendar"
	"github.com/jo-hoe/calendar-manager/internal/ical"
)

type Handler struct {
	provider calendar.Provider
	reader   *ical.Reader
	logger   *slog.Logger
}

func NewHandler(provider calendar.Provider, reader *ical.Reader, logger *slog.Logger) *Handler {
	return &Handler{
		provider: provider,
		reader:   reader,
		logger:   logger,
	}
}

func (h *Handler) HandleProbe(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.reader.FetchEvents(r.Context())
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "failed to fetch events: "+err.Error())
		return
	}

	response := make([]EventResponse, 0, len(events))
	for _, e := range events {
		response = append(response, toEventResponse(e))
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	event, err := toCalendarEvent(req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := h.reader.FindMatch(r.Context(), event.Title, event.StartTime)
	if err != nil {
		h.logger.Warn("failed to check for existing event, proceeding with create", "error", err)
	}

	if existing != nil {
		if err := h.provider.UpdateEvent(r.Context(), existing.ID, event); err != nil {
			h.writeError(w, http.StatusBadGateway, "failed to update event: "+err.Error())
			return
		}
		event.ID = existing.ID
		h.logger.Info("updated existing event", "id", existing.ID, "title", event.Title)
		h.writeJSON(w, http.StatusOK, toEventResponse(event))
		return
	}

	id, err := h.provider.CreateEvent(r.Context(), event)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "failed to create event: "+err.Error())
		return
	}

	event.ID = id
	h.logger.Info("created new event", "id", id, "title", event.Title)
	h.writeJSON(w, http.StatusCreated, toEventResponse(event))
}

func (h *Handler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "event id is required")
		return
	}

	if err := h.provider.DeleteEvent(r.Context(), id); err != nil {
		h.writeError(w, http.StatusBadGateway, "failed to delete event: "+err.Error())
		return
	}

	h.logger.Info("deleted event", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, ErrorResponse{Error: message})
}

func toCalendarEvent(req CreateEventRequest) (calendar.Event, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return calendar.Event{}, err
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return calendar.Event{}, err
	}

	return calendar.Event{
		Title:       req.Title,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    req.Location,
		Description: req.Description,
	}, nil
}

func toEventResponse(e calendar.Event) EventResponse {
	return EventResponse{
		ID:          e.ID,
		Title:       e.Title,
		StartTime:   e.StartTime.Format(time.RFC3339),
		EndTime:     e.EndTime.Format(time.RFC3339),
		Location:    e.Location,
		Description: e.Description,
	}
}

// Shutdown helper for graceful server termination.
func GracefulShutdown(ctx context.Context, server *http.Server, logger *slog.Logger) {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger.Info("shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
}
