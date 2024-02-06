package models

import (
	"context"
	"encoding/json"
	"time"

	"dario.cat/mergo"
	"github.com/ryanfaerman/netctl/internal/dao"
)

type Membership struct {
	Account *Account `json:"account"`
	Target  *Account `json:"target"`
	Role    *Role    `json:"role" validate:"required"`

	CreatedAt time.Time
	ID        int64
}

func FindMemberships(ctx context.Context, account *Account, kind AccountKind) ([]*Membership, error) {
	var memberships []*Membership
	raws, err := global.dao.GetAccountKindMemberships(ctx, dao.GetAccountKindMembershipsParams{
		AccountID: account.ID,
		Kind:      int64(kind),
	})
	if err != nil {
		return memberships, err
	}
	for _, raw := range raws {
		target := &Account{
			ID:        raw.ID,
			Name:      raw.Name,
			About:     raw.About,
			CreatedAt: raw.Createdat,
			Kind:      AccountKind(raw.Kind),
		}
		if err := json.Unmarshal([]byte(raw.Settings), &target.Settings); err != nil {
			return memberships, err
		}

		if err := mergo.Merge(&target.Settings, DefaultSettings); err != nil {
			return memberships, err
		}

		if raw.Deletedat.Valid {
			target.DeletedAt = raw.Deletedat.Time
			target.Deleted = true
		}

		role := &Role{
			ID:          raw.RoleID,
			Name:        raw.RoleName,
			Permissions: Permission(raw.RolePermissions),
			Ranking:     raw.RoleRanking,
		}

		memberships = append(memberships, &Membership{
			Account:   account,
			Target:    target,
			Role:      role,
			CreatedAt: raw.MembershipCreatedAt,
			ID:        raw.ID,
		})
	}
	return memberships, nil
}
