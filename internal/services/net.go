package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/ryanfaerman/netctl/hamdb"
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

func (net) Create(ctx context.Context, name string) (*models.Net, error) {
	id, err := global.dao.CreateNetAndReturnId(ctx, name)
	if err != nil {
		return nil, err
	}
	return models.FindNetById(ctx, id)
}

var CheckinErrClubClass = errors.New("club class")

func (n net) Checkin(ctx context.Context, stream string, checkin *models.NetCheckin) error {
	id := ulid.Make()
	defer func() {
		go func() {
			ctx := context.Background()
			license, err := hamdb.Lookup(ctx, checkin.Callsign.AsHeard)
			if err != nil {
				if err != hamdb.ErrNotFound {
					global.log.Error("hamdb lookup failed", "error", err)
				}
				Event.Create(ctx, stream, events.NetCheckinVerified{
					ErrorType: fmt.Sprintf("%T", err),
				})
				return
			}

			var logicErr error
			if license.Class == hamdb.ClubClass {
				logicErr = CheckinErrClubClass
			}

			Event.Create(ctx, stream, events.NetCheckinVerified{
				ID: id.String(),

				Callsign:  strings.ToUpper(license.Call),
				Name:      license.FullName(),
				Location:  strings.Join([]string{license.City, license.State}, ", "),
				ErrorType: fmt.Sprintf("%T", logicErr),
			})
		}()
	}()

	return Event.Create(ctx, stream, events.NetCheckinHeard{
		ID: id.String(),

		Callsign: strings.ToUpper(checkin.Callsign.AsHeard),
		Name:     checkin.Name.AsHeard,
		Location: checkin.Location.AsHeard,
		Kind:     checkin.Kind.String(),
		Traffic:  0,
	})

	return nil
}

func (s net) GetReplayed(ctx context.Context, id int64) (*models.Net, error) {
	m, err := s.Get(ctx, id)
	if err != nil {
		return m, err
	}
	return m, m.Replay(ctx)
}
