package models

import "time"

type Event[K any] struct {
	ID         int64
	At         time.Time
	StreamID   string
	Originator int64
	Name       string
	Event      K
}

type NonEvent struct {
	ID         int64
	At         time.Time
	StreamID   string
	Originator int64
	Name       string
	Data       []byte
}

type EventStream []Event[any]

func (es EventStream) ForStream(streamID string) EventStream {
	var filtered EventStream
	for _, event := range es {
		if event.StreamID == streamID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
