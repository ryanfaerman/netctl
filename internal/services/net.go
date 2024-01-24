package services

import (
	"context"
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

func (n net) Checkin(ctx context.Context, stream string, checkin *models.NetCheckin) error {
	id := ulid.Make()
	defer func() {
		go func() {
			ctx := context.Background()
			errorType := events.ErrorTypeNone

			license, err := hamdb.Lookup(ctx, checkin.Callsign.AsHeard)
			if err != nil {
				if err == hamdb.ErrNotFound {
					errorType = events.ErrorTypeNotFound
				} else {
					errorType = events.ErrorTypeLookupFailed
					global.log.Error("hamdb lookup failed", "error", err)

				}
				Event.Create(ctx, stream, events.NetCheckinVerified{
					ID:        id.String(),
					ErrorType: errorType.Error(),
				})
				return
			}

			if license.Class == hamdb.ClubClass {
				errorType = events.ErrorTypeClubClass
			}

			Event.Create(ctx, stream, events.NetCheckinVerified{
				ID: id.String(),

				Callsign:  strings.ToUpper(license.Call),
				Name:      license.FullName(),
				Location:  strings.Join([]string{license.City, license.State}, ", "),
				ErrorType: errorType.Error(),
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
}

func (s net) GetReplayed(ctx context.Context, id int64) (*models.Net, error) {
	m, err := s.Get(ctx, id)
	if err != nil {
		return m, err
	}
	return m, m.Replay(ctx)
}
