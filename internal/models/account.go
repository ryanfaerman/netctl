package models

import (
	"context"
	"errors"
	"time"
)

type AccountKind int

//go:generate stringer -type=AccountKind -trimprefix=AccountKind
const (
	AccountKindUser AccountKind = iota
	AccountKindClub
)

var AccountAnonymous = &Account{
	ID:        -1,
	Name:      "Anonymous",
	CreatedAt: time.Now(),
}

type Account struct {
	ID int64 `validate:"gte=0"`

	Name  string `validate:"required"`
	About string
	Kind  AccountKind

	CreatedAt time.Time
	DeletedAt time.Time
	Deleted   bool
}

func (m *Account) IsAnonymous() bool {
	return m.ID < 0
}

func (m *Account) InsertAllowed() bool {
	return !m.IsAnonymous()
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

func FindAccountByID(ctx context.Context, id int64) (*Account, error) {
	raw, err := global.dao.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	u := Account{
		ID:        raw.ID,
		Name:      raw.Name,
		About:     raw.About,
		Kind:      AccountKind(raw.Kind),
		CreatedAt: raw.Createdat,
	}
	if raw.Deletedat.Valid {
		u.DeletedAt = raw.Deletedat.Time
		u.Deleted = true
	}
	return &u, nil
}

func FindAccountByEmail(ctx context.Context, email string) (*Account, error) {
	raw, err := global.dao.FindAccountByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	u := Account{
		ID:        raw.ID,
		Name:      raw.Name,
		About:     raw.About,
		Kind:      AccountKind(raw.Kind),
		CreatedAt: raw.Createdat,
	}
	if raw.Deletedat.Valid {
		u.DeletedAt = raw.Deletedat.Time
		u.Deleted = true
	}
	return &u, nil
}

func FindAccountByCallsign(ctx context.Context, callsign string) (*Account, error) {
	raw, err := global.dao.FindAccountByCallsign(ctx, callsign)
	if err != nil {
		return nil, err
	}
	u := Account{
		ID:        raw.ID,
		Name:      raw.Name,
		About:     raw.About,
		Kind:      AccountKind(raw.Kind),
		CreatedAt: raw.Createdat,
	}
	if raw.Deletedat.Valid {
		u.DeletedAt = raw.Deletedat.Time
		u.Deleted = true
	}
	return &u, nil
}

func (u *Account) Callsigns() ([]Callsign, error) {
	var callsigns []Callsign
	rows, err := global.dao.FindCallsignsForAccount(context.Background(), u.ID)
	if err != nil {
		return callsigns, err
	}
	for _, row := range rows {
		callsign := Callsign{
			ID:   row.ID,
			Call: row.Callsign,
		}
		callsigns = append(callsigns, callsign)
	}
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
