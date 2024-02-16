package finders

import (
	"context"
	"fmt"
	"time"

	ttlcache "github.com/jellydator/ttlcache/v3"
)

// Finder is an interface for finding things. Anything that wishes to be
// findable, must implement this interface. The return type should always
// be a slice a pointer to the type that implements this interface. This allows
// for the consumer to choose between FindOne and Find, without changing
// the interface implementation.
type Finder interface {
	Find(context.Context, QuerySet) (any, error)
}

type FinderCacher interface {
	Finder
	FindCacheKey() string
}

type FinderCacheDurationer interface {
	FinderCacher
	FindCacheDuration() time.Duration
}

type FinderCacheCapacitor interface {
	FinderCacher
	FindCacheCapacity() uint64
}

type FinderCacherDurationerCapacitor interface {
	FinderCacheDurationer
	FinderCacheCapacitor
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

var caches = make(map[string]*ttlcache.Cache[string, any])

const (
	DefaultCacheTTL      = 5 * time.Minute
	DefaultCacheCapacity = 1000
)

func FindCached[K FinderCacher](ctx context.Context, queries ...QueryFunc) ([]*K, error) {
	k := *new(K)
	cacheKey := k.FindCacheKey()

	cache, ok := caches[cacheKey]
	if !ok {
		var maybe any
		maybe = k
		switch T := maybe.(type) {
		case FinderCacherDurationerCapacitor:
			fmt.Println("CacheDurationerCapacitor")
			cache = ttlcache.New[string, any](
				ttlcache.WithTTL[string, any](T.FindCacheDuration()),
				ttlcache.WithCapacity[string, any](T.FindCacheCapacity()),
			)
		case FinderCacheCapacitor:
			fmt.Println("CacheCapacitor")
			cache = ttlcache.New[string, any](
				ttlcache.WithTTL[string, any](DefaultCacheTTL),
				ttlcache.WithCapacity[string, any](T.FindCacheCapacity()),
			)
		case FinderCacheDurationer:
			fmt.Println("CacheDurationer")
			cache = ttlcache.New[string, any](
				ttlcache.WithTTL[string, any](T.FindCacheDuration()),
			)
		default:
			fmt.Println("Default")
			cache = ttlcache.New[string, any](
				ttlcache.WithTTL[string, any](DefaultCacheTTL),
			)
		}
		go cache.Start()
		caches[cacheKey] = cache
	}

	var qs QuerySet
	for _, q := range queries {
		v, err := q()
		if err != nil {
			return nil, err
		}
		qs = append(qs, v)
	}
	qsKey := qs.String()
	if cache.Has(qsKey) {
		return cache.Get(qsKey).Value().([]*K), nil
	}

	results, err := Find[K](ctx, queries...)
	if err != nil {
		return nil, nil
	}

	cache.Set(qsKey, results, ttlcache.DefaultTTL)

	return results, nil
}

func FindOneCached[K FinderCacher](ctx context.Context, queries ...QueryFunc) (*K, error) {
	results, err := FindCached[K](ctx, queries...)
	if err != nil {
		return nil, err
	}
	if len(results) >= 1 {
		return results[0], nil
	}
	return nil, ErrNotFound
}

func ClearFinderCache[K FinderCacher]() {
	k := *new(K)
	cacheKey := k.FindCacheKey()
	if cache, ok := caches[cacheKey]; ok {
		cache.DeleteAll()
	}
}

func EnforceValue[K any](queries QuerySet, key string) (K, error) {
	k := *new(K)
	v, ok := queries.ValueForField(key).(K)
	if !ok {
		return k, ErrInvalidFieldType
	}
	return v, nil
}
