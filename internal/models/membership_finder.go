package models

import (
	"context"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models/finders"
)

// Find[Membership](ctx, ByAccount(1), ByMember(of), WithPermission(PermissionEdit))
func (m Membership) Find(ctx context.Context, queries finders.QuerySet) (any, error) {
	var (
		raw   dao.Membership
		raws  []dao.Membership
		err   error
		found []*Membership
	)

	switch {
	default:
		return nil, finders.ErrInvalidWhere
	case queries.HasWhere("account_id", "member_of") && queries.HasField("permission"):
		memberOf, err := finders.EnforceValue[int64](queries, "member_of")
		if err != nil {
			return nil, err
		}
		accountID, err := finders.EnforceValue[int64](queries, "account_id")
		if err != nil {
			return nil, err
		}
		permission, err := finders.EnforceValue[int64](queries, "permission")
		if err != nil {
			return nil, err
		}

		raw, err = global.dao.HasPermissionOnAccount(ctx, dao.HasPermissionOnAccountParams{
			AccountID:  accountID,
			MemberOf:   memberOf,
			Permission: permission,
		})
		raws = append(raws, raw)
	}
	if err != nil {
		return nil, err
	}

	found = make([]*Membership, len(raws))
	for i, raw := range raws {
		found[i] = &Membership{
			Account: &Account{
				ID: raw.AccountID,
			},
			Target: &Account{
				ID: raw.MemberOf,
			},
			Role: &Role{
				ID: raw.RoleID,
			},
			CreatedAt: raw.CreatedAt,
			ID:        raw.ID,
		}
	}

	return found, nil
}
