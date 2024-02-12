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
		memberOf, ok := finders.EnforceValue[int64](queries, "member_of")
		if ok != nil {
			return nil, ok
		}
		accountID, ok := finders.EnforceValue[int64](queries, "account_id")
		if ok != nil {
			return nil, ok
		}
		permission, ok := finders.EnforceValue[int64](queries, "permission")
		if ok != nil {
			return nil, ok
		}

		raw, err = global.dao.HasPermissionOnAccount(ctx, dao.HasPermissionOnAccountParams{
			AccountID:  accountID,
			MemberOf:   memberOf,
			Permission: permission,
		})
		raws = append(raws, raw)
	case queries.HasWhere("account_id", "kind"):
		accountID, ok := finders.EnforceValue[int64](queries, "account_id")
		if ok != nil {
			return nil, ok
		}
		kind, ok := finders.EnforceValue[int](queries, "kind")
		if ok != nil {
			return nil, ok
		}

		raws, err = global.dao.GetMembershipsForAccountAndKind(ctx, dao.GetMembershipsForAccountAndKindParams{
			AccountID: accountID,
			Kind:      int64(kind),
		})
	case queries.HasWhere("account_id"):
		accountID, ok := finders.EnforceValue[int64](queries, "account_id")
		if ok != nil {
			return nil, ok
		}
		raws, err = global.dao.GetMembershipsForAccount(ctx, accountID)

	}

	if err != nil {
		return nil, err
	}

	found = make([]*Membership, len(raws))
	for i, raw := range raws {
		found[i] = &Membership{
			AccountID: raw.AccountID,
			TargetID:  raw.MemberOf,
			RoleID:    raw.RoleID,
			CreatedAt: raw.CreatedAt,
			ID:        raw.ID,
		}
	}

	return found, nil
}
