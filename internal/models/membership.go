package models

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/ryanfaerman/netctl/internal/models/finders"
)

type Membership struct {
	AccountID int64

	TargetID int64
	RoleID   int64

	CreatedAt time.Time
	ID        int64
}

func (m *Membership) Account(ctx context.Context) *Account {
	account, err := finders.FindOneCached[Account](ctx, finders.ByID(m.AccountID))
	if err != nil {
		fmt.Println("error finding account", "error", err)
		return nil
	}
	return account
}

func (m *Membership) Target(ctx context.Context) *Account {
	a, err := finders.FindOne[Account](ctx, finders.ByID(m.TargetID))
	if err != nil {
		global.log.Error("error finding account", "error", err)
		return AccountAnonymous
	}
	return a
}

func (m *Membership) Role(ctx context.Context) *Role {
	r, err := finders.FindOne[Role](ctx, finders.ByID(m.RoleID))
	if err != nil {
		return RoleNone
	}
	return r
}

func (m *Membership) Can(ctx context.Context, account *Account, action string) error {
	p := ParsePermission(action)

	if !m.Role(ctx).Permissions.Has(p) {
		return fmt.Errorf("account %d does not have permission %s", m.AccountID, p)
	}

	return nil
}

func FindMemberships(ctx context.Context, account *Account, kind AccountKind) ([]*Membership, error) {
	_, file, line, _ := runtime.Caller(1)
	global.log.Warn(
		"DEPRECATED FUNCTION",
		"func", "FindMemberships",
		"use", "Find[Membership](ctx, ByAccount(ID)",
		"file", file,
		"line", line,
	)
	return finders.Find[Membership](ctx, finders.ByAccount(account.ID), finders.ByKind(int(kind)))
}
