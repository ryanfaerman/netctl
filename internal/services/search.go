package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"dario.cat/mergo"
	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models"
)

type search struct {
	userGeoFunc func(ctx context.Context) (float64, float64, error)
}

var Search = search{
	userGeoFunc: func(ctx context.Context) (float64, float64, error) {
		account := Session.GetAccount(ctx)
		return Account.Geolocation(ctx, account)
	},
}

/*
*
*
* SELECT id, latitude, longitude,
    6371 * 2 * ASIN(SQRT(POWER(SIN((?1 - ABS(latitude)) * pi()/180 / 2), 2) +
        COS(?1 * pi()/180 ) * COS(ABS(latitude) * pi()/180) *
        POWER(SIN((?2 - longitude) * pi()/180 / 2), 2))) AS distance
FROM locations
WHERE distance <= ?3;
* */

const callsignsWithinRange = `-- name: CallsignsWithinRange :many
SELECT
  accounts.id, accounts.createdat, accounts.updatedat, 
  accounts.deletedat, accounts.kind, 
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

type CallsignsWithinRangeParams struct {
	Latitude  float64
	Longitude float64
	Distance  float64
	Kind      int
}

func (s search) CallsignsWithinRange(ctx context.Context, arg CallsignsWithinRangeParams) ([]*models.Account, error) {
	rows, err := global.db.QueryContext(ctx, callsignsWithinRange, arg.Latitude, arg.Longitude, arg.Distance, arg.Kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Account
	for rows.Next() {
		var (
			raw dao.Account
			i   models.Account
		)

		err := rows.Scan(
			&raw.ID,
			&raw.Createdat,
			&raw.Updatedat,
			&raw.Deletedat,
			&raw.Kind,
			&raw.Settings,
			&raw.Slug,
			&i.Distance,
		)
		if err != nil {
			return nil, err
		}

		i.ID = raw.ID
		i.Slug = raw.Slug
		i.Kind = models.AccountKind(raw.Kind)
		i.CreatedAt = raw.Createdat
		if raw.Deletedat.Valid {
			i.DeletedAt = raw.Deletedat.Time
			i.Deleted = true
		}
		if err := json.Unmarshal([]byte(raw.Settings), &i.Settings); err != nil {
			fmt.Println("error unmarshalling settings", err)
			return nil, err
		}

		if err := mergo.Merge(&i.Settings, models.DefaultSettings); err != nil {
			return nil, err
		}

		i.Name = i.Settings.Name
		i.About = i.Settings.About

		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// near:me -- search for accounts near the user
// near:GridSquare -- search for accounts near a grid square
// within:range of:gridquare -- search for accounts within a range of a grid square
// within:range of:LAT,LON -- search for accounts within a range of a lat/lon
// kind:KIND -- search for accounts of a specific kind
// callsign:NAME -- search for accounts with a specific name
type Filter struct {
	Range     float64
	Latitude  float64
	Longitude float64
	Kind      models.AccountKind
	Callsign  string
	Term      string
}

var operators = map[string]func(...string) Filter{
	"near": func(args ...string) Filter {
		if len(args) == 1 && args[0] == "me" {
			lat, lon, err := Search.userGeoFunc(context.Background())
			if err != nil {
				panic(err.Error())
			}
			return Filter{
				Range:     10,
				Latitude:  lat,
				Longitude: lon,
			}
		}
		if len(args) == 2 {
			lat, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				panic(err.Error())
			}
			lon, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				panic(err.Error())
			}
			return Filter{
				Latitude:  lat,
				Longitude: lon,
				Range:     10,
			}
		}
		panic("invalid arguments for near")
		return Filter{}
	},
	"within": func(args ...string) Filter {
		if len(args) != 1 {
			panic("expected 1 argument for within")
		}
		distance, unit, err := parseDistance(args[0])
		if err != nil {
			panic(err.Error())
		}
		if strings.HasPrefix(unit, "m") {
			distance = distance * 1.60934
		}

		return Filter{
			Range: distance,
		}
	},
	"of": func(args ...string) Filter {
		if len(args) != 2 {
			panic("expected 2 arguments for of")
		}

		lat, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			panic(err.Error())
		}
		lon, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			panic(err.Error())
		}
		return Filter{
			Latitude:  lat,
			Longitude: lon,
		}
	},
	"kind": func(args ...string) Filter {
		if len(args) != 1 {
			panic("expected 1 argument for kind")
		}
		return Filter{
			Kind: models.ParseAccountKind(args[0]),
		}
	},
}

func (s search) ParseQuery(raw string) Filter {
	raw = strings.ToLower(raw)
	f := Filter{
		Kind: models.AccoundKindAny,
	}

	terms := []string{}
	for _, part := range strings.Fields(raw) {
		fmt.Println(part)
		parts := strings.SplitN(part, ":", 2)
		if len(parts) != 2 {
			terms = append(terms, part)
			continue
		}

		op, ok := operators[parts[0]]
		if !ok {
			terms = append(terms, part)
			continue
		}

		frag := op(strings.Split(parts[1], ",")...)

		if err := mergo.Merge(&f, frag, mergo.WithOverride); err != nil {
			panic(err.Error())
		}
		spew.Dump(frag)

		spew.Dump(parts)

	}
	f.Term = strings.Join(terms, " ")

	return f
}

func parseDistance(distance string) (float64, string, error) {
	// Remove all spaces and trim any leading/trailing spaces
	distance = strings.TrimSpace(strings.ReplaceAll(distance, " ", ""))

	// Find the index where the unit starts
	unitIndex := len(distance)
	for i := 0; i < len(distance); i++ {
		if !isDigit(distance[i]) {
			unitIndex = i
			break
		}
	}

	// Extract the value and unit parts
	valuePart := distance[:unitIndex]
	unitPart := distance[unitIndex:]

	// Convert the value part to a float64
	value, err := strconv.ParseFloat(valuePart, 64)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse value: %v", err)
	}

	// Return the value and unit parts
	return value, unitPart, nil
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9' || char == '.'
}
