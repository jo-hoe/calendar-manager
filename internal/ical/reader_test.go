package ical

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testICalData = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Google Inc//Google Calendar 70.9054//EN
BEGIN:VEVENT
DTSTART;TZID=Europe/Berlin:20260510T160000
DTEND;TZID=Europe/Berlin:20260510T180000
SUMMARY:Union Berlin Frauen vs TSG Hoffenheim
LOCATION:STADION An der Alten Försterei\, Berlin
DESCRIPTION:Frauen-Bundesliga | 25. Spieltag
UID:abc123@google.com
END:VEVENT
BEGIN:VEVENT
DTSTART:20260502T133000Z
DTEND:20260502T153000Z
SUMMARY:Union Berlin vs 1. FC Köln
LOCATION:Stadion An der Alten Försterei
UID:def456@google.com
END:VEVENT
END:VCALENDAR`

func TestFetchEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(testICalData))
	}))
	defer server.Close()

	reader := NewReader(server.URL)
	events, err := reader.FetchEvents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Title != "Union Berlin Frauen vs TSG Hoffenheim" {
		t.Errorf("expected title 'Union Berlin Frauen vs TSG Hoffenheim', got %q", events[0].Title)
	}
	if events[0].ID != "abc123" {
		t.Errorf("expected ID 'abc123', got %q", events[0].ID)
	}
	if events[0].Location != "STADION An der Alten Försterei, Berlin" {
		t.Errorf("unexpected location: %q", events[0].Location)
	}
}

func TestFindMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(testICalData))
	}))
	defer server.Close()

	reader := NewReader(server.URL)

	loc, _ := time.LoadLocation("Europe/Berlin")
	startTime := time.Date(2026, 5, 10, 16, 0, 0, 0, loc)

	match, err := reader.FindMatch(context.Background(), "Union Berlin Frauen vs TSG Hoffenheim", startTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match == nil {
		t.Fatal("expected a match, got nil")
	}
	if match.ID != "abc123" {
		t.Errorf("expected ID 'abc123', got %q", match.ID)
	}
}

func TestFindMatchNoMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(testICalData))
	}))
	defer server.Close()

	reader := NewReader(server.URL)

	match, err := reader.FindMatch(context.Background(), "Nonexistent Event", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match != nil {
		t.Errorf("expected nil, got %v", match)
	}
}

func TestFetchEventsEmptyCalendar(t *testing.T) {
	emptyICal := "BEGIN:VCALENDAR\nVERSION:2.0\nEND:VCALENDAR"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(emptyICal))
	}))
	defer server.Close()

	reader := NewReader(server.URL)
	events, err := reader.FetchEvents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}
