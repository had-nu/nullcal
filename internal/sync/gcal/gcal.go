package gcal

import (
	"context"
	"fmt"
	"time"

	googlecalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/had-nu/nullcal/internal/store"
)

// Adapter implements a read-only Google Calendar integration.
// Write operations (Create, Update, Delete) are no-ops in Phase 1.
type Adapter struct {
	svc *googlecalendar.Service
}

// New creates a new GCal Adapter.
// It performs the OAuth2 flow on first use and caches the token.
func New(ctx context.Context) (*Adapter, error) {
	client, err := AuthClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcal auth: %w", err)
	}
	svc, err := googlecalendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("gcal service: %w", err)
	}
	return &Adapter{svc: svc}, nil
}

// ListEvents fetches events from the primary Google Calendar between from and to.
func (a *Adapter) ListEvents(ctx context.Context, from, to time.Time) ([]store.CalendarEvent, error) {
	res, err := a.svc.Events.List("primary").
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("gcal list events: %w", err)
	}

	events := make([]store.CalendarEvent, 0, len(res.Items))
	for _, item := range res.Items {
		start, err := parseGCalTime(item.Start)
		if err != nil {
			continue
		}
		end, err := parseGCalTime(item.End)
		if err != nil {
			continue
		}
		events = append(events, store.CalendarEvent{
			ExternalID:  item.Id,
			Source:      "gcal",
			Title:       item.Summary,
			StartAt:     start,
			EndAt:       end,
			Description: item.Description,
			SyncedAt:    time.Now(),
		})
	}
	return events, nil
}

// parseGCalTime converts a GCal EventDateTime (supports both dateTime and date-only).
func parseGCalTime(dt *googlecalendar.EventDateTime) (time.Time, error) {
	if dt == nil {
		return time.Time{}, fmt.Errorf("nil EventDateTime")
	}
	if dt.DateTime != "" {
		return time.Parse(time.RFC3339, dt.DateTime)
	}
	// All-day events use date-only format.
	return time.Parse("2006-01-02", dt.Date)
}

// toGCalEventDateTime converts a time.Time to a GCal EventDateTime.
// If the time is midnight exactly, we treat it as an all-day event.
func toGCalEventDateTime(t time.Time) *googlecalendar.EventDateTime {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return &googlecalendar.EventDateTime{Date: t.Format("2006-01-02")}
	}
	return &googlecalendar.EventDateTime{DateTime: t.Format(time.RFC3339)}
}

// CreateEvent creates a new event in the primary Google Calendar.
// If the task has no time (midnight), an all-day event is created.
func (a *Adapter) CreateEvent(ctx context.Context, title, description string, startAt, endAt time.Time) (string, error) {
	ev := &googlecalendar.Event{
		Summary:     title,
		Description: description,
		Start:       toGCalEventDateTime(startAt),
		End:         toGCalEventDateTime(endAt),
	}
	created, err := a.svc.Events.Insert("primary", ev).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gcal create event: %w", err)
	}
	return created.Id, nil
}

// UpdateEvent updates an existing GCal event.
func (a *Adapter) UpdateEvent(ctx context.Context, eventID, title, description string, startAt, endAt time.Time) error {
	ev := &googlecalendar.Event{
		Summary:     title,
		Description: description,
		Start:       toGCalEventDateTime(startAt),
		End:         toGCalEventDateTime(endAt),
	}
	_, err := a.svc.Events.Update("primary", eventID, ev).Context(ctx).Do()
	return err
}

// DeleteEvent removes a GCal event.
func (a *Adapter) DeleteEvent(ctx context.Context, eventID string) error {
	return a.svc.Events.Delete("primary", eventID).Context(ctx).Do()
}

