package models

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/ryanfaerman/netctl/internal/events"
)

type Net struct {
	ID   int64
	Name string

	Sessions map[string]*NetSession
}

func (n Net) String() string {

	var b strings.Builder

	b.WriteString(fmt.Sprintf("Net %d: %s\n", n.ID, n.Name))
	b.WriteString("---\n")
	b.WriteString("Sessions: \n")

	for _, session := range n.Sessions {
		b.WriteString(fmt.Sprintf("  ID(%s):\n", session.ID))
		b.WriteString(fmt.Sprintf("    Status: %s\n", session.Status.String()))
		b.WriteString("    Checkins: \n")
		for _, checkin := range session.Checkins {
			b.WriteString(fmt.Sprintf("    - Callsign: %s\n", checkin.Callsign))
			b.WriteString(fmt.Sprintf("      Kind: %s\n", checkin.Kind.String()))
			b.WriteString(fmt.Sprintf("      Acked: %t\n", checkin.Acked))
			b.WriteString(fmt.Sprintf("      Time: %s\n", checkin.At))
		}
	}

	return b.String()

}

func (n *Net) Events(ctx context.Context) (EventStream, error) {
	events, err := global.dao.GetNetEvents(ctx, n.ID)
	if err != nil {
		return nil, err
	}

	stream := make(EventStream, len(events))

	for i, raw := range events {
		decoder := gob.NewDecoder(bytes.NewReader(raw.EventData))
		var p any
		if err := decoder.Decode(&p); err != nil {
			global.log.Error("unable to decode event", "error", err)
			// panic(err)
			return stream, err
		}
		stream[i] = Event[any]{
			ID:         raw.ID,
			At:         raw.Created,
			StreamID:   raw.SessionID,
			Originator: raw.AccountID,
			Name:       raw.EventType,
			Event:      p,
		}
	}
	return stream, nil
}

func (n *Net) Replay(ctx context.Context) error {
	stream, err := n.Events(ctx)
	if err != nil {
		return err
	}
	for _, event := range stream {
		switch e := event.Event.(type) {
		case events.NetStarted:
			n.Sessions[event.StreamID] = &NetSession{
				ID:     event.StreamID,
				Status: NetStatusOpened,
			}
		case events.NetCheckin:
			c := NetCheckin{
				Callsign: e.Callsign,
				At:       event.At,
				Kind:     NetCheckinKindRoutine,
			}
			n.Sessions[event.StreamID].Checkins = append(n.Sessions[event.StreamID].Checkins, c)
		case events.NetAckCheckin:
			session := n.Sessions[event.StreamID]
			for i, checkin := range session.Checkins {
				if checkin.Callsign == e.Callsign {
					session.Checkins[i].Acked = true
				}
			}
		}
	}
	return nil
}

//go:generate stringer -type=NetStatus -trimprefix=NetStatus
type NetStatus int

const (
	NetStatusUnknown NetStatus = iota
	NetStatusScheduled
	NetStatusOpened
	NetStatusClosed
)

type NetSession struct {
	ID     string
	Status NetStatus

	Checkins []NetCheckin
}

//go:generate stringer -type=NetCheckinKind -trimprefix=NetCheckinKind
type NetCheckinKind int

const (
	NetCheckinKindUnknown NetCheckinKind = iota
	NetCheckinKindRoutine
	NetCheckinKindPriority
	NetCheckinKindTraffic
)

type NetCheckin struct {
	Callsign string
	At       time.Time
	Kind     NetCheckinKind
	Acked    bool
}
