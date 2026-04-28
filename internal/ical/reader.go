package ical

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/jo-hoe/calendar-manager/internal/calendar"
)

type Reader struct {
	url        string
	httpClient *http.Client
}

func NewReader(icalURL string) *Reader {
	return &Reader{
		url:        icalURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *Reader) FetchEvents(ctx context.Context) ([]calendar.Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching ical: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from ical url", resp.StatusCode)
	}

	return parseICalBody(resp.Body)
}

func (r *Reader) FindMatch(ctx context.Context, title string, startTime time.Time) (*calendar.Event, error) {
	events, err := r.FetchEvents(ctx)
	if err != nil {
		return nil, err
	}

	for i := range events {
		if events[i].Title == title && events[i].StartTime.Equal(startTime) {
			return &events[i], nil
		}
	}

	return nil, nil
}

func parseICalBody(body io.Reader) ([]calendar.Event, error) {
	cal, err := ics.ParseCalendar(body)
	if err != nil {
		return nil, fmt.Errorf("parsing ical: %w", err)
	}

	var events []calendar.Event
	for _, component := range cal.Events() {
		event, err := convertVEvent(component)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func convertVEvent(vevent *ics.VEvent) (calendar.Event, error) {
	startTime, err := parseEventTime(vevent, ics.ComponentPropertyDtStart)
	if err != nil {
		return calendar.Event{}, fmt.Errorf("parsing start time: %w", err)
	}

	endTime, err := parseEventTime(vevent, ics.ComponentPropertyDtEnd)
	if err != nil {
		endTime = startTime.Add(time.Hour)
	}

	uid := vevent.GetProperty(ics.ComponentPropertyUniqueId)
	eventID := ""
	if uid != nil {
		eventID = stripGoogleSuffix(uid.Value)
	}

	return calendar.Event{
		ID:          eventID,
		Title:       getPropertyValue(vevent, ics.ComponentPropertySummary),
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    getPropertyValue(vevent, ics.ComponentPropertyLocation),
		Description: getPropertyValue(vevent, ics.ComponentPropertyDescription),
	}, nil
}

func parseEventTime(vevent *ics.VEvent, prop ics.ComponentProperty) (time.Time, error) {
	p := vevent.GetProperty(prop)
	if p == nil {
		return time.Time{}, fmt.Errorf("property %s not found", prop)
	}

	tzID := p.ICalParameters["TZID"]
	value := p.Value

	if len(tzID) > 0 && tzID[0] != "" {
		loc, err := time.LoadLocation(tzID[0])
		if err == nil {
			t, err := time.ParseInLocation("20060102T150405", value, loc)
			if err == nil {
				return t, nil
			}
		}
	}

	for _, layout := range []string{
		time.RFC3339,
		"20060102T150405Z",
		"20060102T150405",
		"20060102",
	} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time value: %s", value)
}

func getPropertyValue(vevent *ics.VEvent, prop ics.ComponentProperty) string {
	p := vevent.GetProperty(prop)
	if p == nil {
		return ""
	}
	return p.Value
}

func stripGoogleSuffix(uid string) string {
	return strings.TrimSuffix(uid, "@google.com")
}
