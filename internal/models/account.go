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
	ID   int64
	Slug string `form:"slug" json:"slug"`

	Name  string `form:"name" validate:"required"`
	About string `form:"about" json:"about"`
	Kind  AccountKind

	Settings Settings

	CreatedAt time.Time
	DeletedAt time.Time
	Deleted   bool

	Distance float64

	callsigns []Callsign
}

func init() {
	gob.Register(Account{})
}

func (m *Account) Verbs() []string {
	return []string{
		"edit", "view", "view-location", "view-activity-graph",
	}
}

func (m *Account) Can(account *Account, action string) error {
	switch action {
	case "edit":
		if account.IsAnonymous() {
			return errors.New("anonymous users cannot edit accounts")
		}
		if account.ID != m.ID {
			return errors.New("cannot edit another user's account")
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

func (u *Account) Emails() ([]Email, error) {
	var emails []Email
	rows, err := global.dao.GetEmailsForAccount(context.Background(), u.ID)
	if err != nil {
		return emails, err
	}
	for _, row := range rows {
		email := Email{
			ID:           row.ID,
			CreatedAt:    row.Createdat,
			UpdatedAt:    row.Updatedat,
			Address:      row.Address,
			IsPrimary:    row.Isprimary,
			IsPublic:     row.Ispublic,
			IsNotifiable: row.Isnotifiable,
		}

		if row.Verifiedat.Valid {
			email.VerifiedAt = row.Verifiedat.Time
			email.IsVerified = true
		}

		emails = append(emails, email)
	}
	return emails, nil
}

func (m *Account) PrimaryEmail() (Email, error) {
	emails, err := m.Emails()
	if err != nil {
		return Email{}, err
	}
	for _, email := range emails {
		if email.IsPrimary {
			return email, nil
		}
	}
	return Email{}, errors.New("no primary email")
}

func (m *Account) Clubs(ctx context.Context) ([]*Membership, error) {
	return FindMemberships(ctx, m, AccountKindClub)
}

func (m *Account) Organizations(ctx context.Context) ([]*Membership, error) {
	return FindMemberships(ctx, m, AccountKindOrganization)
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

func (u *Account) Callsigns() ([]Callsign, error) {
	if len(u.callsigns) > 0 {
		fmt.Println("using preloaded callsigns")
		return u.callsigns, nil
	}
	var callsigns []Callsign
	rows, err := global.dao.FindCallsignsForAccount(context.Background(), u.ID)
	if err != nil {
		return callsigns, err
	}
	for _, row := range rows {
		callsign := Callsign{
			ID:         row.ID,
			Call:       row.Callsign,
			Expires:    row.Expires.Time,
			Status:     row.Status,
			Latitude:   row.Latitude.Float64,
			Longitude:  row.Longitude.Float64,
			Firstname:  row.Firstname.String,
			Middlename: row.Middlename.String,
			Lastname:   row.Lastname.String,
			Suffix:     row.Suffix.String,
			Address:    row.Address.String,
			City:       row.City.String,
			State:      row.State.String,
			Zip:        row.Zip.String,
			Country:    row.Country.String,
		}
		callsigns = append(callsigns, callsign)
	}
	u.callsigns = callsigns

	return callsigns, nil
}

func (m *Account) Callsign() Callsign {
	calls, err := m.Callsigns()
	if err != nil {
		return Callsign{}
	}
	if len(calls) == 0 {
		return Callsign{}
	}

	return calls[0]
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
