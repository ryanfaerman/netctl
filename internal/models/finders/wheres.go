package finders

func ByEmail(email string) QueryFunc {
	return func() (Query, error) {
		return Query{Type: QueryWhere, Fields: []string{"email"}, Values: []any{email}}, nil
	}
}

func ByCallsign(callsign string) QueryFunc {
	return func() (Query, error) {
		return Query{Type: QueryWhere, Fields: []string{"callsign"}, Values: []any{callsign}}, nil
	}
}

func ByID(id int64) QueryFunc {
	return func() (Query, error) {
		return Query{Type: QueryWhere, Fields: []string{"id"}, Values: []any{id}}, nil
	}
}

func ByStreamID(stream string) QueryFunc {
	return func() (Query, error) {
		return Query{Type: QueryWhere, Fields: []string{"stream_id"}, Values: []any{stream}}, nil
	}
}

func ByDistance(lat, lon, distance float64) QueryFunc {
	return func() (Query, error) {
		return Query{
			Type:   QueryWhere,
			Fields: []string{"distance"},
			Values: []any{lat, lon, distance},
		}, nil
	}
}

func ByKind(kind int) QueryFunc {
	return func() (Query, error) {
		return Query{Type: QueryWhere, Fields: []string{"kind"}, Values: []any{kind}}, nil
	}
}
