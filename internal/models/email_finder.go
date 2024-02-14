package models

import (
	"context"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models/finders"
)

func (m Email) Find(ctx context.Context, queries finders.QuerySet) (any, error) {
	var (
		raw   dao.Email
		raws  []dao.Email
		err   error
		found []*Email
	)

	switch {
	default:
		return nil, finders.ErrInvalidWhere
	case queries.HasWhere("account_id"):
		accountID, ok := finders.EnforceValue[int64](queries, "account_id")
		if ok != nil {
			return nil, ok
		}

		raws, err = global.dao.GetEmailsForAccount(ctx, accountID)
	case queries.HasWhere("id"):
		id, ok := finders.EnforceValue[int64](queries, "id")
		if ok != nil {
			return nil, ok
		}

		raw, err = global.dao.GetEmail(ctx, id)
		raws = append(raws, raw)
	}
	if err != nil {
		return nil, err
	}

	found = make([]*Email, len(raws))
	for i, raw := range raws {
		found[i] = &Email{
			ID:        raw.ID,
			CreatedAt: raw.Createdat,
			Address:   raw.Address,
		}

		if raw.Verifiedat.Valid {
			found[i].VerifiedAt = raw.Verifiedat.Time
			found[i].IsVerified = true
		}

	}
	return found, nil
}
