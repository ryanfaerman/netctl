package finders

import (
	"context"
)

// Finder is an interface for finding things. Anything that wishes to be
// findable, must implement this interface. The return type should always
// be a slice a pointer to the type that implements this interface. This allows
// for the consumer to choose between FindOne and Find, without changing
// the interface implementation.
type Finder interface {
	Find(context.Context, QuerySet) (any, error)
}

// Find the all instances of a Finder type, given a set of queries.
func Find[K Finder](ctx context.Context, queries ...QueryFunc) ([]*K, error) {
	k := *new(K)

	var qs QuerySet
	for _, q := range queries {
		v, err := q()
		if err != nil {
			return nil, err
		}
		qs = append(qs, v)
	}

	if len(qs.OfType(QueryWhere)) == 0 {
		return nil, ErrMissingWhere
	}

	results, err := k.Find(ctx, qs)
	if err != nil {
		return nil, err
	}

	return results.([]*K), nil
}

// FindOne finds a single instance of a Finder type, given a set of queries.
func FindOne[K Finder](ctx context.Context, queries ...QueryFunc) (*K, error) {
	results, err := Find[K](ctx, queries...)
	if err != nil {
		return nil, err
	}
	if len(results) >= 1 {
		return results[0], nil
	}

	return nil, ErrNotFound
}
