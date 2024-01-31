package models

import (
	"context"
	"sort"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/ryanfaerman/netctl/internal/events"
)

type Net struct {
	Name       string                 `validate:"required" json:"name"`
	StreamID   string                 `json:"stream_id"`
	Sessions   map[string]*NetSession `json:"sessions"`
	SessionIDs []string               `json:"session_ids"`
	ID         int64                  `json:"id"`
}

func NewNet(name string) *Net {
	return &Net{
		Name:     name,
		StreamID: ulid.Make().String(),
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
			StreamID: raw.StreamID,
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
		StreamID: raw.StreamID,
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
		m.SessionIDs = append(m.SessionIDs, raw.StreamID)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(m.SessionIDs)))

	return m, nil
}

func FindNetByStreamID(ctx context.Context, streamID string) (*Net, error) {
	raw, err := global.dao.GetNetByStreamID(ctx, streamID)
	if err != nil {
		global.log.Error("cannot execute query", "query", "GetNetByStreamID", "streamID", streamID, "err", err)
		return nil, err
	}
	m := &Net{
		ID:       raw.ID,
		Name:     raw.Name,
		StreamID: raw.StreamID,
		Sessions: make(map[string]*NetSession),
	}
	raws, err := global.dao.GetNetSessions(ctx, raw.ID)
	if err != nil {
		global.log.Error("cannot execute query", "query", "GetNetSessions", "streamID", streamID, "err", err)
		return nil, err

	}
	for _, raw := range raws {
		m.Sessions[raw.StreamID] = &NetSession{
			ID:        raw.StreamID,
			CreatedAt: raw.Created,
		}
		m.SessionIDs = append(m.SessionIDs, raw.StreamID)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(m.SessionIDs)))

	return m, nil
}

func FindNetBySessionID(ctx context.Context, sessionID string) (*Net, error) {
	raw, err := global.dao.GetNetForSession(ctx, sessionID)
	if err != nil {
		global.log.Error("cannot execute query", "query", "GetNetBySessionID", "id", sessionID, "err", err)
		return nil, err
	}

	m := &Net{
		ID:       raw.ID,
		Name:     raw.Name,
		StreamID: raw.StreamID,
		Sessions: make(map[string]*NetSession),
	}

	m.Sessions[sessionID] = &NetSession{
		ID:        sessionID,
		CreatedAt: raw.SessionCreated,
	}
	return m, nil
}

func (n *Net) Events(ctx context.Context, onlyStreams ...string) (EventStream, error) {
	if len(onlyStreams) == 0 {
		streamIDs := make([]string, 0, len(n.Sessions))
		for streamID := range n.Sessions {
			streamIDs = append(streamIDs, streamID)
		}
		if len(streamIDs) == 0 {
			return EventStream{}, nil
		}
		return FindEventsForStreams(ctx, streamIDs...)
	}
	return FindEventsForStreams(ctx, onlyStreams...)
}

func (m *Net) Replay(ctx context.Context, onlyStreams ...string) error {
	stream, err := m.Events(ctx, onlyStreams...)
	if err != nil {
		return err
	}
	m.replay(stream)

	return nil
}

// replay the given event stream, vivifying the model
func (m *Net) replay(stream EventStream) {
	for _, event := range stream {
	eventMachine:
		switch e := event.Event.(type) {
		case *events.NetSessionScheduled:
			// if any periods exist, ignore
			// otherwise, create a new one in the future
			session := m.Sessions[event.StreamID]
			if len(session.Periods) > 0 {
				break eventMachine
			}
			session.Periods = append(session.Periods, Period{
				OpenedAt:  e.At,
				Scheduled: true,
			})

		case *events.NetSessionOpened:
			session := m.Sessions[event.StreamID]

			// if no periods exist, create a new one
			if len(session.Periods) == 0 {
				session.Periods = append(session.Periods, Period{
					OpenedAt: event.At,
				})
				break eventMachine
			}

			// if the last period is scheduled, close it and create a new one
			if session.Periods[len(session.Periods)-1].Scheduled {
				session.Periods[len(session.Periods)-1].ClosedAt = event.At
				session.Periods = append(session.Periods, Period{
					OpenedAt: event.At,
				})
				break eventMachine
			}

			// if the last period is open, ignore
			if session.Periods[len(session.Periods)-1].IsClosed() {
				session.Periods = append(session.Periods, Period{
					OpenedAt: event.At,
				})
				break eventMachine
			}

		case *events.NetSessionClosed:
			// if no periods exist, ignore
			// if the last period is open, close it
			// if the last period is closed, ignore
			session := m.Sessions[event.StreamID]
			if len(session.Periods) == 0 {
				break eventMachine
			}
			if session.Periods[len(session.Periods)-1].IsOpen() {
				session.Periods[len(session.Periods)-1].ClosedAt = event.At
			}

		case *events.NetCheckinHeard:
			session := m.Sessions[event.StreamID]
			// if the checkin is not in the session, add it
			// if the checkin is in the session, reset it
			for i, checkin := range session.Checkins {
				if checkin.ID == e.ID || strings.EqualFold(checkin.Callsign.AsHeard, e.Callsign) {
					session.Checkins[i].Acked = false
					// session.Checkins[i].Verified = false
					// session.Checkins[i].Valid = nil
					break eventMachine
				}
			}
			session.Checkins = append(session.Checkins, NetCheckin{
				ID:       e.ID,
				Callsign: Hearable{AsHeard: e.Callsign},
				Location: Hearable{AsHeard: e.Location},
				Name:     Hearable{AsHeard: e.Name},
				Kind:     ParseNetCheckinKind(e.Kind),
				Traffic:  e.Traffic,
				At:       event.At,
			})

		case *events.NetCheckinVerified:
			// set the verified flag to true
			// if the verification has no errors, set the valid flag to nil
			// if the verification has an error, set the valid flag to the error
			session := m.Sessions[event.StreamID]
			for i, checkin := range session.Checkins {
				if checkin.ID == e.ID {
					session.Checkins[i].Verified = true
					session.Checkins[i].Callsign.AsLicensed = e.Callsign
					session.Checkins[i].Name.AsLicensed = e.Name
					session.Checkins[i].Location.AsLicensed = e.Location

					switch e.ErrorType {
					case "", events.ErrorTypeNone.Error():
						session.Checkins[i].Valid = nil
					case events.ErrorTypeNotFound.Error():
						session.Checkins[i].Valid = events.ErrorTypeNotFound
					case events.ErrorTypeClubClass.Error():
						session.Checkins[i].Valid = events.ErrorTypeClubClass
					case events.ErrorTypeLookupFailed.Error():
						session.Checkins[i].Valid = events.ErrorTypeLookupFailed
					default:
						session.Checkins[i].Valid = events.ErrorTypeLookupFailed
						global.log.Warn(
							"unknown checkin validation error",
							"error", e.ErrorType,
							"event", event.Name,
							"event.id", event.ID,
						)
					}

					break eventMachine
				}
			}
		case *events.NetCheckinAcked:
			// set the acked flag to true
			session := m.Sessions[event.StreamID]
			for i, checkin := range session.Checkins {
				if checkin.ID == e.ID {
					session.Checkins[i].Acked = true
					break eventMachine
				}
			}
		case *events.NetCheckinCorrected:
			// find the checkin and update it
			// mark as invalidated
			session := m.Sessions[event.StreamID]
			// if the checkin is not in the session, add it
			// if the checkin is in the session, reset it
			for i, checkin := range session.Checkins {
				if checkin.ID == e.ID {
					session.Checkins[i].Verified = false
					session.Checkins[i].Valid = nil

					session.Checkins[i].Callsign = Hearable{AsHeard: e.Callsign}
					session.Checkins[i].Location = Hearable{AsHeard: e.Location}
					session.Checkins[i].Name = Hearable{AsHeard: e.Name}
					session.Checkins[i].Kind = ParseNetCheckinKind(e.Kind)
					session.Checkins[i].Traffic = e.Traffic

					break eventMachine
				}
			}
		case *events.NetCheckinRevoked:
			// find the checkin and remove it
			session := m.Sessions[event.StreamID]
			for i, checkin := range session.Checkins {
				if checkin.ID == e.ID {
					session.Checkins = append(session.Checkins[:i], session.Checkins[i+1:]...)
					break eventMachine
				}
			}
		}
	}
}
