package models

import (
	"context"
	"time"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/events"
)

type Event struct {
	Event     any
	At        time.Time
	Name      string
	StreamID  string
	ID        int64
	AccountID int64
}

// FindEventsForStreams returns a stream of events for the given streamIDs.
func FindEventsForStreams(ctx context.Context, streamIDs ...string) (EventStream, error) {
	if len(streamIDs) == 0 {
		return nil, nil
	}

	raws, err := global.dao.GetEventsForStreams(ctx, streamIDs)
	if err != nil {
		global.log.Error("unable to get events for streams", "error", err, "streams", streamIDs)
		return nil, err
	}

	stream := make(EventStream, len(raws))

	for i, raw := range raws {
		e, err := events.Decode(raw.EventType, []byte(raw.EventData))
		if err != nil {
			return EventStream{}, err
		}

		stream[i] = Event{
			ID:        raw.ID,
			At:        raw.Created,
			StreamID:  raw.StreamID,
			AccountID: raw.AccountID,
			Name:      raw.EventType,
			Event:     e,
		}
	}

	return stream, nil
}

// FindEventsForCallsign returns a stream of events for the given callsign and event type.
func FindEventsForCallsign(eventType string, callsign string) (EventStream, error) {
	l := global.log.With("callsign", callsign, "event_type", eventType)
	raws, err := global.dao.GetEventsForCallsign(context.Background(), dao.GetEventsForCallsignParams{
		EventType: eventType,
		Callsign:  callsign,
	})
	if err != nil {
		l.Error("unable to get events for callsign")
		return nil, err
	}

	stream := make(EventStream, len(raws))

	for i, raw := range raws {
		e, err := events.Decode(raw.EventType, []byte(raw.EventData))
		if err != nil {
			return EventStream{}, err
		}

		stream[i] = Event{
			ID:        raw.ID,
			At:        raw.Created,
			StreamID:  raw.StreamID,
			AccountID: raw.AccountID,
			Name:      raw.EventType,
			Event:     e,
		}
	}

	return stream, nil
}

type RecoveredEvent struct {
	RegisteredFn string
	SessionToken string
	Event        Event
	ID           int64
}

// FindRecoverableEvents returns a stream of events that have been registered for recovery.
func FindRecoverableEvents(ctx context.Context) ([]RecoveredEvent, error) {
	raws, err := global.dao.GetRecoverableEvents(ctx)
	if err != nil {
		global.log.Error("unable to get recoverable events", "error", err)
		return nil, err
	}
	stream := make([]RecoveredEvent, len(raws))

	for i, raw := range raws {
		e, err := events.Decode(raw.EventType, []byte(raw.EventData))
		if err != nil {
			return stream, err
		}

		stream[i] = RecoveredEvent{
			ID:           raw.RecoveryID,
			RegisteredFn: raw.RegisteredFn,
			SessionToken: raw.SessionToken,
			Event: Event{
				ID:        raw.ID,
				At:        raw.Created,
				StreamID:  raw.StreamID,
				AccountID: raw.AccountID,
				Name:      raw.EventType,
				Event:     e,
			},
		}
	}
	return stream, nil
}

// An EventStream is a stream of events.
type EventStream []Event

func (es EventStream) FilterForStream(streamID string) EventStream {
	var filtered EventStream
	for _, event := range es {
		if event.StreamID == streamID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func (es EventStream) FilterForName(name string) EventStream {
	var filtered EventStream
	for _, event := range es {
		if event.Name == name {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
