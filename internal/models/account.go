package models

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ryanfaerman/netctl/internal/dao"

	. "github.com/ryanfaerman/netctl/internal/models/finders"
)

type AccountKind int

//go:generate stringer -type=AccountKind --linecomment
const (
	AccountKindUser         AccountKind = iota // user
	AccountKindClub                            // club
	AccountKindOrganization                    // organization
	AccoundKindAny                             // any
)

func ParseAccountKind(s string) AccountKind {
	switch strings.ToLower(s) {
	case "club":
		return AccountKindClub
	case "organization":
		return AccountKindOrganization
	default:
		return AccountKindUser
	}
}

var AccountAnonymous = &Account{
	ID:        -1,
	Name:      "Anonymous",
	CreatedAt: time.Now(),
}

type Account struct {
	ID       int64
	Slug     string `form:"slug" json:"slug"`
	StreamID string `json:"stream_id"`

	Name  string `form:"name" json:"name" validate:"required"`
	About string `form:"about" json:"about"`
	Kind  AccountKind

	Settings Settings

	CreatedAt time.Time
	DeletedAt time.Time
	Deleted   bool

	Distance float64

	callsigns []Callsign

	Cached bool
}

func init() {
	gob.Register(Account{})
}

func (m *Account) Verbs() []string {
	return []string{
		"edit", "view", "view-location", "view-activity-graph",
	}
}

func (m *Account) Can(ctx context.Context, account *Account, action string) error {
	switch action {
	case "edit":
		if account.IsAnonymous() {
			return errors.New("anonymous users cannot edit accounts")
		}
		if account.ID != m.ID {
			memberships, err := Find[Membership](ctx, ByAccount(account.ID), ByMemberOf(m.ID), WithPermission(int64(PermissionEdit)))
			if err != nil {
				return fmt.Errorf("unable to check permission: %w", err)
			}
			if len(memberships) == 0 {
				return errors.New("cannot edit another user's account")
			}

		}
	case "view":
		// if account.IsAnonymous() {
		// 	return errors.New("account is restricted")
		// }
		return nil
	case "view-location":
		switch m.Settings.PrivacySettings.Location {
		case "private":
			return errors.New("location viewing is prohibited")
		case "protected":
			if account.IsAnonymous() {
				return errors.New("location viewing is prohibited")
			}
		}
	case "view-activity-graph":
		if m.Settings.AppearanceSettings.ActivityGraphs == "off" {
			return errors.New("activity graphs are disabled")
		}

	}

	return nil
}

func (m *Account) IsAnonymous() bool {
	return m.ID < 0
}

func (m *Account) InsertAllowed() bool {
	return !m.IsAnonymous()
}

func (m *Account) Setting(ctx context.Context, path string) any {
	if !strings.HasPrefix(path, ".") {
		path = fmt.Sprintf(".%s", path)
	}

	l := global.log.With("account_id", m.ID, "path", path)
	l.Warn("getting settings")
	raw, err := global.dao.GetAccountSetting(ctx, dao.GetAccountSettingParams{
		ID:       m.ID,
		Jsonpath: fmt.Sprintf("$%s", strings.ToLower(path)),
	})
	if err != nil {
		l.Error("unable to get account setting", "error", err)
		panic(err.Error())
	}
	if raw == nil {
		panic("raw is nil")
	}

	return raw
}

func (u *Account) Emails() ([]*Email, error) {
	return Find[Email](context.Background(), ByAccount(u.ID))
}

func (m *Account) PrimaryEmail() Email {
	emails, err := m.Emails()
	if err != nil {
		return Email{}
	}
	if len(emails) > 0 {
		return *emails[0]
	}
	return Email{}
}

func (m *Account) Members(ctx context.Context) []*Membership {
	members, err := Find[Membership](ctx, ByMemberOf(m.ID))
	if err != nil {
		return []*Membership{}
	}
	return members
}

func (m *Account) Delegated(ctx context.Context) ([]*Membership, error) {
	return Find[Membership](ctx, ByAccount(m.ID))
}

func (m *Account) Clubs(ctx context.Context) ([]*Membership, error) {
	return Find[Membership](ctx, ByAccount(m.ID), ByKind(int(AccountKindClub)))
}

func (m *Account) Organizations(ctx context.Context) ([]*Membership, error) {
	return Find[Membership](ctx, ByAccount(m.ID), ByKind(int(AccountKindOrganization)))
}

func FindAccountByID(ctx context.Context, id int64) (*Account, error) {
	_, file, line, _ := runtime.Caller(1)
	global.log.Warn(
		"DEPRECATED FUNCTION",
		"func", "FindAccountByID",
		"use", "FindOne[Account](ctx, ByID(id))",
		"file", file,
		"line", line,
	)

	return FindOne[Account](ctx, ByID(id))
}

func FindAccountByEmail(ctx context.Context, email string) (*Account, error) {
	_, file, line, _ := runtime.Caller(1)
	global.log.Warn(
		"DEPRECATED FUNCTION",
		"func", "FindAccountByEmail",
		"use", "FindOne[Account](ctx, ByEmail(email))",
		"file", file,
		"line", line,
	)

	return FindOne[Account](ctx, ByEmail(email))
}

func FindAccountByCallsign(ctx context.Context, callsign string) (*Account, error) {
	_, file, line, _ := runtime.Caller(1)
	global.log.Warn(
		"DEPRECATED FUNCTION",
		"func", "FindAccountByCallsign",
		"use", "FindOne[Account](ctx, ByCallsign(email))",
		"file", file,
		"line", line,
	)

	return FindOne[Account](ctx, ByCallsign(callsign))
}

func (u *Account) Callsigns(ctx context.Context) ([]*Callsign, error) {
	return Find[Callsign](ctx, ByAccount(u.ID))
}

func (m *Account) Callsign(ctx context.Context) *Callsign {
	calls, err := m.Callsigns(ctx)
	if err != nil {
		return &Callsign{}
	}
	if len(calls) == 0 {
		return &Callsign{}
	}

	return calls[0]
}

func (m *Account) Location(ctx context.Context) (float64, float64) {
	if m.Settings.LocationSettings.HasLocation() {
		return m.Settings.LocationSettings.Location()
	}
	callsign := m.Callsign(ctx)
	return callsign.Latitude, callsign.Longitude
}

var (
	ErrAccountNeedsCallsign = errors.New("user needs a callsign")
	ErrAccountNeedsName     = errors.New("user needs a name")
)

func (m *Account) Ready() error {
	var errs []error

	{
		rows, err := global.dao.FindCallsignsForAccount(context.Background(), m.ID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			errs = append(errs, ErrAccountNeedsCallsign)
		}
	}

	{
		if m.Name == "" {
			errs = append(errs, ErrAccountNeedsName)
		}
	}

	return errors.Join(errs...)
}
