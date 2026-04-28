package calendar

import "context"

type Provider interface {
	CreateEvent(ctx context.Context, event Event) (string, error)
	UpdateEvent(ctx context.Context, eventID string, event Event) error
	DeleteEvent(ctx context.Context, eventID string) error
}
