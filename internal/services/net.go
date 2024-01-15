package services

import (
	"context"

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

// func (n net) StartSession(ctx context.Context, id int64) error {
//
//	}
func (n net) Checkin(ctx context.Context, stream string, checkin *models.NetCheckin) error {
	return Event.Create(ctx, stream, events.NetCheckin{
		Callsign: checkin.Callsign,
	})
}

func (s net) GetReplayed(ctx context.Context, id int64) (*models.Net, error) {
	m, err := s.Get(ctx, id)
	if err != nil {
		return m, err
	}
	return m, m.Replay(ctx)
}

// func (net) SaveEventForNet(id int64, stream string, e any) error {
// 	var b bytes.Buffer
// 	var p any
// 	p = &e
// 	if err := gob.NewEncoder(&b).Encode(p); err != nil {
// 		return err
// 	}
//
// 	return global.dao.CreateNetEvent(context.Background(), dao.CreateNetEventParams{
// 		NetID:     id,
// 		SessioID: stream,
// 		AccountID: 1,
// 		EventType: fmt.Sprintf("%T", e),
// 		EventData: b.Bytes(),
// 	})
//
// 	return nil
// }
