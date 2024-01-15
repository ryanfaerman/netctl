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

type Account struct {
	ID int64

	Name string
	Kind AccountKind

	CreatedAt time.Time
	DeletedAt time.Time
	Deleted   bool
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

func FindAccountByID(ctx context.Context, id int64) (*Account, error) {
	raw, err := global.dao.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	u := Account{
		ID:        raw.ID,
		Name:      raw.Name,
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
