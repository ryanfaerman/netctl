package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/r3labs/sse/v2"
	"github.com/ryanfaerman/netctl/hamdb"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/events"
	"github.com/ryanfaerman/netctl/internal/models"
)

type net struct{}

var Net net

func (net) All(ctx context.Context) ([]*models.Net, error) {
	return models.FindAllNets(ctx)
}

func (net) Get(ctx context.Context, id int64) (*models.Net, error) {
	return models.FindNetById(ctx, id)
}

func (net) GetByStreamID(ctx context.Context, streamID string) (*models.Net, error) {
	return models.FindNetByStreamID(ctx, streamID)
}

// CreateSession creates a new session and associates it with the net.
func (net) CreateSession(ctx context.Context, netStreamID string) (*models.NetSession, error) {
	session_id := ulid.Make().String()

	net, err := models.FindNetByStreamID(ctx, netStreamID)
	if err != nil {
		return nil, err
	}

	_, err = global.dao.CreateNetSessionAndReturnId(ctx, dao.CreateNetSessionAndReturnIdParams{
		NetID:    net.ID,
		StreamID: session_id,
	})
	if err != nil {
		return nil, err
	}
	return &models.NetSession{
		ID: session_id,
	}, nil
}

func (net) GetNetFromSession(ctx context.Context, sessionID string) (*models.Net, error) {
	n, err := models.FindNetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return n, n.Replay(ctx, sessionID)
}

// func (net) Create(ctx context.Context, name string) (*models.Net, error) {
// 	id, err := global.dao.CreateNetAndReturnId(ctx, name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return models.FindNetById(ctx, id)
// }

func (net) Create(ctx context.Context, m *models.Net) (*models.Net, error) {
	if err := Validation.Apply(m); err != nil {
		return m, err
	}
	m.StreamID = ulid.Make().String()

	id, err := global.dao.CreateNetAndReturnId(ctx, dao.CreateNetAndReturnIdParams{
		Name:     m.Name,
		StreamID: m.StreamID,
	})
	if err != nil {
		return m, err
	}
	m.ID = id

	return m, err
}

func init() {
	Breaker.AddWithConsecutive("hamdb", 5) // TODO: make this configurable
	Event.Register(events.NetCheckinHeard{}, Net.ValidateCheckinHeard)
}

func (n net) ValidateCheckinHeard(ctx context.Context, event models.Event) error {
	if e, ok := event.Event.(events.NetCheckinHeard); ok {
		m := &models.NetCheckin{
			ID: e.ID,
		}
		m.Callsign.AsHeard = e.Callsign

		global.log.Debug("validating checkin heard", "callsign", m.Callsign.AsHeard)
		return Breaker.Call("hamdb", func() error {
			return n.ValidateCheckin(ctx, event.StreamID, m)
		})

	}
	global.log.Debug("invalid event", "event", event)
	return errors.New("invalid event")
}

func (n net) ValidateCheckin(ctx context.Context, stream string, checkin *models.NetCheckin) error {
	license, err := hamdb.Lookup(ctx, checkin.Callsign.AsHeard)
	if err != nil {
		if err == hamdb.ErrNotFound {
			return Event.Create(ctx, stream, events.NetCheckinVerified{
				ID:        checkin.ID,
				ErrorType: events.ErrorTypeNotFound.Error(),
			})
		}
		return err
	}

	errorType := events.ErrorTypeNone
	if license.Class == hamdb.ClubClass {
		errorType = events.ErrorTypeClubClass
	}

	if err := Event.Create(ctx, stream, events.NetCheckinVerified{
		ID: checkin.ID,

		Callsign:  strings.ToUpper(license.Call),
		Name:      license.FullName(),
		Location:  strings.Join([]string{license.City, license.State}, ", "),
		ErrorType: errorType.Error(),
	}); err != nil {
		return err
	}

	net, err := n.GetNetFromSession(ctx, stream)
	if err != nil {
		return err
	}

	existing := net.Sessions[stream].FindCheckinByCallsign(checkin.Callsign.AsHeard)

	if checkin.ID != existing.ID {
		global.log.Info("sending validation SSE event", "id", existing.ID, "existing", true)
		Event.Server.Publish(stream, &sse.Event{
			Event: []byte(existing.ID),
			Data:  []byte("validation"),
		})

		return nil
	}

	global.log.Info("sending validation SSE event", "id", existing.ID, "existing", false)
	Event.Server.Publish(stream, &sse.Event{
		Event: []byte(checkin.ID),
		Data:  []byte("validation"),
	})

	return nil
}

var ErrCheckinExists = errors.New("checkin exists")

func (n net) Checkin(ctx context.Context, stream string, m *models.NetCheckin) (*models.NetCheckin, error) {
	if err := Validation.Apply(m); err != nil {
		return m, err
	}

	m.ID = ulid.Make().String()

	if err := Event.Create(ctx, stream, events.NetCheckinHeard{
		ID: m.ID,

		Callsign: strings.ToUpper(m.Callsign.AsHeard),
		Name:     m.Name.AsHeard,
		Location: m.Location.AsHeard,
		Kind:     m.Kind.String(),
		Traffic:  m.Traffic,
	}); err != nil {
		return m, err
	}
	m.At = time.Now()

	net, err := n.GetNetFromSession(ctx, stream)
	if err != nil {
		return m, err
	}

	existing := net.Sessions[stream].FindCheckinByCallsign(m.Callsign.AsHeard)

	if m.ID != existing.ID {
		global.log.Info("sending validation SSE event", "id", existing.ID, "existing", true)
		Event.Server.Publish(stream, &sse.Event{
			Event: []byte(existing.ID),
			Data:  []byte("nack"),
		})
		return m, ErrCheckinExists

	}

	return m, nil
}

func (s net) GetReplayed(ctx context.Context, id int64) (*models.Net, error) {
	m, err := s.Get(ctx, id)
	if err != nil {
		return m, err
	}
	return m, m.Replay(ctx)
}

func (s net) AckCheckin(ctx context.Context, stream string, id string) error {
	defer func() {
		global.log.Info("sending validation event", "id", id)
		Event.Server.Publish(stream, &sse.Event{
			Event: []byte(id),
			Data:  []byte("ack"),
		})
	}()
	return Event.Create(ctx, stream, events.NetCheckinAcked{
		ID: id,
	})
}
