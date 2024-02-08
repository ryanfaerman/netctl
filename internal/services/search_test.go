package services

import (
	"context"
	"testing"

	"github.com/ryanfaerman/netctl/internal/models"
)

func TestSearchFilterParsing(t *testing.T) {
	Search.userGeoFunc = func(ctx context.Context) (float64, float64, error) {
		return 26.2711019, -80.2457015, nil
	}

	examples := map[string]struct {
		query  string
		filter Filter
	}{
		"simple text": {
			query:  "foo bar",
			filter: Filter{Term: "foo bar", Kind: models.AccoundKindAny},
		},
		"near me": {
			query:  "near:me",
			filter: Filter{Range: 10, Latitude: 26.2711019, Longitude: -80.2457015, Kind: models.AccoundKindAny},
		},
		"spurious colon": {
			query:  "something: else",
			filter: Filter{Term: "something: else", Kind: models.AccoundKindAny},
		},
		"non-operator": {
			query:  "do:thing",
			filter: Filter{Term: "do:thing", Kind: models.AccoundKindAny},
		},
		"near me within": {
			query:  "near:me within:30mi",
			filter: Filter{Range: 30 * 1.60934, Latitude: 26.2711019, Longitude: -80.2457015, Kind: models.AccoundKindAny},
		},
		"of:LAT,LON": {
			query:  "of:26.2711019,-80.2457015",
			filter: Filter{Latitude: 26.2711019, Longitude: -80.2457015, Kind: models.AccoundKindAny},
		},
		"kind": {
			query:  "kind:club",
			filter: Filter{Kind: models.AccountKindClub},
		},
		"precedence": {
			query:  "kind:club kind:organization",
			filter: Filter{Kind: models.AccountKindOrganization},
		},
	}

	for name, example := range examples {
		name := name
		example := example
		t.Run(name, func(t *testing.T) {
			if filter := Search.ParseQuery(example.query); filter != example.filter {
				t.Errorf("ParseFilter(%s) = %v, expected %v", example.query, filter, example.filter)
			}
		})
	}
}
