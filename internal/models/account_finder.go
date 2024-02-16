package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models/finders"
)

func (m Account) FindCacheKey() string {
	return "accounts"
}

func (m *Account) FindCacheDuration() time.Duration {
	return 10 * time.Minute
}

func (m Account) Find(ctx context.Context, queries finders.QuerySet) (any, error) {
	var (
		raw   dao.Account
		raws  []dao.Account
		err   error
		found []*Account
	)

	switch {
	default:
		return nil, finders.ErrInvalidWhere
	case queries.HasWhere("id"):
		val, ok := queries.ValueForField("id").(int64)
		if !ok {
			return nil, finders.ErrInvalidFieldType
		}
		if val <= 0 {
			return nil, finders.ErrInvalidFieldValue
		}
		raw, err = global.dao.GetAccount(ctx, val)
		raws = append(raws, raw)

	case queries.HasWhere("email"):
		val, ok := queries.ValueForField("email").(string)
		if !ok {
			return nil, finders.ErrInvalidFieldType
		}
		if val == "" {
			return nil, finders.ErrInvalidFieldValue
		}
		raw, err = global.dao.FindAccountByEmail(ctx, val)
		raws = append(raws, raw)

	case queries.HasWhere("callsign"):
		val, ok := queries.ValueForField("callsign").(string)
		if !ok {
			return nil, finders.ErrInvalidFieldType
		}
		if val == "" {
			return nil, finders.ErrInvalidFieldValue
		}
		raw, err = global.dao.FindAccountByCallsign(ctx, val)
		raws = append(raws, raw)
	case queries.HasWhere("slug"):
		val, ok := queries.ValueForField("slug").(string)
		if !ok {
			return nil, finders.ErrInvalidFieldType
		}
		raw, err = global.dao.GetAccountBySlug(ctx, val)
		raws = append(raws, raw)

	case queries.HasWhere("distance", "kind"):
		vals := queries.ValuesForField("distance")
		if len(vals) < 3 {
			return nil, finders.ErrInvalidFieldValue
		}

		lat, ok := vals[0].(float64)
		if !ok {
			return nil, finders.ErrInvalidFieldValue
		}

		lon, ok := vals[1].(float64)
		if !ok {
			return nil, finders.ErrInvalidFieldValue
		}

		distance, ok := vals[2].(float64)
		if !ok {
			return nil, finders.ErrInvalidFieldValue
		}

		kind, ok := queries.ValueForField("kind").(int)
		if !ok {
			return nil, finders.ErrInvalidFieldValue
		}

		r, distances, qerr := callsignsWithinRange(ctx, callsignsWithinRangeParams{
			Latitude:  lat,
			Longitude: lon,
			Distance:  distance,
			Kind:      int(AccountKind(kind)),
		})
		raws = r
		err = qerr

		if qerr == nil {
			defer func() {
				for i := range found {
					found[i].Distance = distances[i]
				}
			}()
		}

	}

	if err != nil {
		return nil, err
	}

	found = make([]*Account, len(raws))

	for i, raw := range raws {
		a := Account{
			ID:        raw.ID,
			Kind:      AccountKind(raw.Kind),
			CreatedAt: raw.Createdat,
			Settings:  DefaultSettings,
			Slug:      raw.Slug,
			StreamID:  raw.StreamID,
		}
		if raw.Deletedat.Valid {
			a.DeletedAt = raw.Deletedat.Time
			a.Deleted = true
		}
		if err := json.Unmarshal([]byte(raw.Settings), &a.Settings); err != nil {
			return nil, err
		}

		a.Name = a.Settings.ProfileSettings.Name
		a.About = a.Settings.ProfileSettings.About

		if !a.Settings.LocationSettings.HasLocation() {
			callsigns, err := a.Callsigns()
			if err == nil {
				if len(callsigns) > 0 {
					callsign := callsigns[0]
					a.Settings.LocationSettings.Latitude = callsign.Latitude
					a.Settings.LocationSettings.Longitude = callsign.Longitude
				}
			}
		}

		switch {
		case queries.HasField("callsigns"):
			if _, err := (&a).Callsigns(); err != nil {
				return nil, err
			}
		}

		found[i] = &a
	}

	return found, nil
}

const callsignsWithinRangeQuery = `-- name: CallsignsWithinRange :many
SELECT
  accounts.id, accounts.name, accounts.createdat, accounts.updatedat, 
  accounts.deletedat, accounts.kind, accounts.about, 
  accounts.settings, accounts.slug,
  6371 * 2 * ASIN(SQRT(POWER(SIN((?1 - ABS(latitude)) * pi()/180 / 2), 2) +
    COS(?1 * pi()/180 ) * COS(ABS(latitude) * pi()/180) *
    POWER(SIN((?2 - longitude) * pi()/180 / 2), 2))) AS distance
FROM callsigns
JOIN accounts_callsigns on accounts_callsigns.callsign_id = callsigns.id
JOIN accounts on accounts_callsigns.account_id = accounts.id
WHERE 
  accounts.kind = ?4
  AND distance > 0
  AND distance <= ?3 
  AND (
    accounts.settings->"$.privacy.visibility"  != 'private'
    OR accounts.settings->"$.privacy.visibility" IS NULL
  )
`

type callsignsWithinRangeParams struct {
	Latitude  float64
	Longitude float64
	Distance  float64
	Kind      int
}

func callsignsWithinRange(ctx context.Context, arg callsignsWithinRangeParams) ([]dao.Account, []float64, error) {
	rows, err := global.db.QueryContext(ctx, callsignsWithinRangeQuery, arg.Latitude, arg.Longitude, arg.Distance, arg.Kind)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var (
		items     []dao.Account
		distances []float64
	)
	for rows.Next() {
		var (
			raw      dao.Account
			distance float64
		)

		err := rows.Scan(
			&raw.ID,
			&raw.Createdat,
			&raw.Updatedat,
			&raw.Deletedat,
			&raw.Kind,
			&raw.Settings,
			&raw.Slug,
			&distance,
		)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, raw)
		distances = append(distances, distance)
	}
	if err := rows.Close(); err != nil {
		return nil, nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, distances, nil
}
