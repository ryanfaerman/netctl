package services

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models"
)

type membership struct{}

var Membership membership

func (s membership) Create(ctx context.Context, owner, m *models.Account) error {
	if owner.IsAnonymous() {
		return fmt.Errorf("anonymous users cannot create organizations")
	}
	if err := Validation.Apply(m); err != nil {
		return err
	}

	tx, err := global.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := global.dao.WithTx(tx)

	id, err := qtx.CreateAccount(ctx, dao.CreateAccountParams{
		Name: m.Name,
		Kind: int64(m.Kind),
		Slug: m.Slug,
	})
	if err != nil {
		return fmt.Errorf("error creating account: %w", err)
	}
	m.ID = id
	spew.Dump("new account", m)

	roleID, err := qtx.CreateRoleOnAccount(ctx, dao.CreateRoleOnAccountParams{
		Name:        "Owner",
		AccountID:   m.ID,
		Permissions: int64(models.PermissionOwner),
		Ranking:     0,
	})
	if err != nil {
		return fmt.Errorf("error creating role: %w", err)
	}
	fmt.Println("roleID", roleID)

	params := dao.CreateMembershipParams{
		AccountID: owner.ID,
		MemberOf:  m.ID,
		RoleID:    roleID,
	}

	spew.Dump(params)
	err = qtx.CreateMembership(ctx, params)
	if err != nil {
		return fmt.Errorf("error creating membership: %w", err)
	}

	return tx.Commit()
}
