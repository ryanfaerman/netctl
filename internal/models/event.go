package models

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"
)

type Event struct {
	ID        int64
	At        time.Time
	StreamID  string
	AccountID int64
	Name      string
	Event     any
}

func FindEventsForStreams(ctx context.Context, streamIDs ...string) (EventStream, error) {
	raws, err := global.dao.GetEventsForStreams(ctx, streamIDs)
	if err != nil {
		global.log.Error("unable to get events for streams", "error", err, "streams", streamIDs)
		return nil, err
	}

	stream := make(EventStream, len(raws))

	for i, raw := range raws {
		decoder := gob.NewDecoder(bytes.NewReader(raw.EventData))
		var p any
		if err := decoder.Decode(&p); err != nil {
			global.log.Error("unable to decode event", "error", err)
			return stream, err
		}

		stream[i] = Event{
			ID:        raw.ID,
			At:        raw.Created,
			StreamID:  raw.StreamID,
			AccountID: raw.AccountID,
			Name:      raw.EventType,
			Event:     p,
		}
	}

	return stream, nil
}

type RecoveredEvent struct {
	Event        Event
	RegisteredFn string
	ID           int64
}

func FindRecoverableEvents(ctx context.Context) ([]RecoveredEvent, error) {
	raws, err := global.dao.GetRecoverableEvents(ctx)
	if err != nil {
		global.log.Error("unable to get recoverable events", "error", err)
		return nil, err
	}
	stream := make([]RecoveredEvent, len(raws))

	for i, raw := range raws {
		decoder := gob.NewDecoder(bytes.NewReader(raw.EventData))
		var p any
		if err := decoder.Decode(&p); err != nil {
			global.log.Error("unable to decode event", "error", err)
			return stream, err
		}

		stream[i] = RecoveredEvent{
			ID:           raw.RecoveryID,
			RegisteredFn: raw.RegisteredFn,
			Event: Event{
				ID:        raw.ID,
				At:        raw.Created,
				StreamID:  raw.StreamID,
				AccountID: raw.AccountID,
				Name:      raw.EventType,
				Event:     p,
			},
		}
	}
	return stream, nil
}

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
