package finders

func WithSettings() (Query, error) {
	return Query{Type: QueryField, Fields: []string{"settings"}, Values: []any{true}}, nil
}

func WithCallsign() (Query, error) {
	return Query{Type: QueryField, Fields: []string{"callsign"}, Values: []any{true}}, nil
}
