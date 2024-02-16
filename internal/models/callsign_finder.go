package models

import (
	"context"
	"time"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models/finders"
)

func (m Callsign) FindCacheKey() string {
	return "callsigns"
}

func (m Callsign) FindCacheDuration() time.Duration {
	return 7 * 24 * time.Hour
}

func (m Callsign) Find(ctx context.Context, queries finders.QuerySet) (any, error) {
	var (
		raws  []dao.Callsign
		err   error
		found []*Callsign
	)

	switch {
	default:
		return nil, finders.ErrInvalidWhere
	case queries.HasWhere("account_id"):
		accountID, ok := finders.EnforceValue[int64](queries, "account_id")
		if ok != nil {
			return nil, ok
		}
		raws, err = global.dao.FindCallsignsForAccount(ctx, accountID)

	}

	if err != nil {
		return nil, err
	}

	found = make([]*Callsign, len(raws))
	for i, raw := range raws {
		found[i] = &Callsign{
			ID:         raw.ID,
			Call:       raw.Callsign,
			Class:      raw.Class,
			Expires:    raw.Expires.Time,
			Status:     raw.Status,
			Latitude:   raw.Latitude.Float64,
			Longitude:  raw.Longitude.Float64,
			Firstname:  raw.Firstname.String,
			Middlename: raw.Middlename.String,
			Lastname:   raw.Lastname.String,
			Suffix:     raw.Suffix.String,
			Address:    raw.Address.String,
			City:       raw.City.String,
			State:      raw.State.String,
			Zip:        raw.Zip.String,
			Country:    raw.Country.String,
		}
	}

	return found, nil
}
