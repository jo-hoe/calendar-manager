package calendar

import "time"

type Event struct {
	ID          string
	Title       string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Description string
}
