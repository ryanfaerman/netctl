package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	ulid "github.com/oklog/ulid/v2"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/events"
)

type Net struct {
	ID   int64
	Name string

	Sessions map[string]*NetSession
}

func NewNet(id int64, name string) *Net {
	return &Net{
		ID:       id,
		Name:     name,
		Sessions: make(map[string]*NetSession),
	}
}

func FindAllNets(ctx context.Context) ([]*Net, error) {
	raws, err := global.dao.GetNets(ctx)
	if err != nil {
		return nil, err
	}

	nets := make([]*Net, len(raws))

	for i, raw := range raws {

		nets[i] = &Net{
			ID:       raw.ID,
			Name:     raw.Name,
			Sessions: make(map[string]*NetSession),
		}
	}

	return nets, nil
}

func FindNetById(ctx context.Context, id int64) (*Net, error) {
	raw, err := global.dao.GetNet(ctx, id)
	if err != nil {
		global.log.Error("cannot execute query", "query", "GetNet", "id", id, "err", err)
		return nil, err
	}
	m := &Net{
		ID:       raw.ID,
		Name:     raw.Name,
		Sessions: make(map[string]*NetSession),
	}
	raws, err := global.dao.GetNetSessions(ctx, id)
	if err != nil {
		global.log.Error("cannot execute query", "query", "GetNetSessions", "id", id, "err", err)
		return nil, err

	}
	for _, raw := range raws {
		m.Sessions[raw.StreamID] = &NetSession{
			ID:        raw.StreamID,
			CreatedAt: raw.Created,
		}
	}
	return m, nil
}

func (m *Net) AddSession(ctx context.Context) (*NetSession, error) {
	streamID := ulid.Make().String()
	session := &NetSession{
		ID: streamID,
	}

	_, err := global.dao.CreateNetSessionAndReturnId(ctx, dao.CreateNetSessionAndReturnIdParams{
		NetID:    m.ID,
		StreamID: streamID,
	})
	if err != nil {
		return nil, err
	}

	m.Sessions[streamID] = session
	return session, nil
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
	streamIDs := make([]string, 0, len(n.Sessions))
	for streamID := range n.Sessions {
		streamIDs = append(streamIDs, streamID)
	}
	if len(streamIDs) == 0 {
		return EventStream{}, nil
	}

	return FindEventsForStreams(ctx, streamIDs...)
}

func (n *Net) Replay(ctx context.Context) error {
	stream, err := n.Events(ctx)
	if err != nil {
		return err
	}
	for _, event := range stream {
		switch e := event.Event.(type) {
		case events.NetStarted:
			n.Sessions[event.StreamID].Status = NetStatusOpened
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
	ID        string
	CreatedAt time.Time
	Status    NetStatus

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
