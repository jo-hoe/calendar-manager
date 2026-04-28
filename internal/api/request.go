package api

import (
	"errors"
	"time"
)

type CreateEventRequest struct {
	Title       string `json:"title"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type EventResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (r CreateEventRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.StartTime == "" {
		return errors.New("startTime is required")
	}
	if r.EndTime == "" {
		return errors.New("endTime is required")
	}

	if _, err := time.Parse(time.RFC3339, r.StartTime); err != nil {
		return errors.New("startTime must be in RFC3339 format (e.g. 2026-05-10T16:00:00+02:00)")
	}
	if _, err := time.Parse(time.RFC3339, r.EndTime); err != nil {
		return errors.New("endTime must be in RFC3339 format (e.g. 2026-05-10T18:00:00+02:00)")
	}

	return nil
}
