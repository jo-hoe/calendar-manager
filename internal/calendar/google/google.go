package google

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jo-hoe/calendar-manager/internal/calendar"
	"github.com/jo-hoe/calendar-manager/internal/config"
	"golang.org/x/oauth2/google"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Client struct {
	service    *gcal.Service
	calendarID string
}

func New(cfg config.GoogleConfig) (*Client, error) {
	credBytes, err := os.ReadFile(cfg.CredentialsFile) // #nosec G304 -- credentials path from config
	if err != nil {
		return nil, fmt.Errorf("reading credentials file: %w", err)
	}

	jwtConfig, err := google.JWTConfigFromJSON(credBytes, gcal.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	ctx := context.Background()
	httpClient := jwtConfig.Client(ctx)

	service, err := gcal.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("creating calendar service: %w", err)
	}

	return &Client{
		service:    service,
		calendarID: cfg.CalendarID,
	}, nil
}

func (c *Client) CreateEvent(ctx context.Context, event calendar.Event) (string, error) {
	gcalEvent := toGoogleEvent(event)

	created, err := c.service.Events.Insert(c.calendarID, gcalEvent).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("inserting event: %w", err)
	}

	return created.Id, nil
}

func (c *Client) UpdateEvent(ctx context.Context, eventID string, event calendar.Event) error {
	gcalEvent := toGoogleEvent(event)

	_, err := c.service.Events.Update(c.calendarID, eventID, gcalEvent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating event %s: %w", eventID, err)
	}

	return nil
}

func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	err := c.service.Events.Delete(c.calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting event %s: %w", eventID, err)
	}

	return nil
}

func toGoogleEvent(event calendar.Event) *gcal.Event {
	return &gcal.Event{
		Summary:     event.Title,
		Location:    event.Location,
		Description: event.Description,
		Start: &gcal.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
			TimeZone: event.StartTime.Location().String(),
		},
		End: &gcal.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
			TimeZone: event.EndTime.Location().String(),
		},
	}
}
