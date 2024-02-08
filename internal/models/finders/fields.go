package finders

func WithSettings() (Query, error) {
	return Query{Type: QueryField, Fields: []string{"settings"}, Values: []any{true}}, nil
}

func WithCallsigns() (Query, error) {
	return Query{Type: QueryField, Fields: []string{"callsigns"}, Values: []any{true}}, nil
}
