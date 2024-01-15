package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/events"
	"github.com/ryanfaerman/netctl/internal/models"
)

type net struct{}

var Net net

func (net) Get(ctx context.Context, id int64) (*models.Net, error) {
	raw, err := global.dao.GetNet(ctx, id)
	if err != nil {
		return nil, err
	}

	m := &models.Net{
		ID:   raw.ID,
		Name: raw.Name,
	}
	return m, nil

}

func (n net) StartSession(ctx context.Context, id int64) error {

	return n.SaveEventForNet(id, "asdf", events.NetStarted{})
}

func (n net) Checkin(ctx context.Context, id int64) error {
	return n.SaveEventForNet(id, "asdf", events.NetCheckin{
		Callsign: "W1AW",
	})
}

func (net) GetReplayed(ctx context.Context, id int64) (*models.Net, error) {
	raw, err := global.dao.GetNet(ctx, id)
	if err != nil {
		return nil, err
	}
	m := &models.Net{
		ID:       raw.ID,
		Name:     raw.Name,
		Sessions: make(map[string]*models.NetSession),
	}
	return m, m.Replay(ctx)
}

func (net) SaveEventForNet(id int64, stream string, e any) error {
	var b bytes.Buffer
	var p any
	p = &e
	if err := gob.NewEncoder(&b).Encode(p); err != nil {
		return err
	}

	return global.dao.CreateNetEvent(context.Background(), dao.CreateNetEventParams{
		NetID:     id,
		SessionID: stream,
		AccountID: 1,
		EventType: fmt.Sprintf("%T", e),
		EventData: b.Bytes(),
	})

	return nil
}
