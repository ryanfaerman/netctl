package models

import (
	"context"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models/finders"
)

func (m Role) FindCacheKey() string {
	return "roles"
}

func (m Role) Find(ctx context.Context, queries finders.QuerySet) (any, error) {
	var (
		raw   dao.Role
		raws  []dao.Role
		err   error
		found []*Role
	)
	switch {
	default:
		return nil, finders.ErrInvalidWhere
	case queries.HasWhere("id"):
		id, ok := finders.EnforceValue[int64](queries, "id")
		if ok != nil {
			return nil, ok
		}

		raw, err = global.dao.GetRole(ctx, id)
		raws = append(raws, raw)
	}

	if err != nil {
		return nil, err
	}

	found = make([]*Role, len(raws))

	for i, raw := range raws {
		found[i] = &Role{
			ID:          raw.ID,
			Name:        raw.Name,
			Permissions: Permission(raw.Permissions),
			Ranking:     raw.Ranking,
		}
	}

	return found, nil
}
