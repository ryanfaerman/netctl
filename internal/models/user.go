package models

import (
	"context"
	"errors"
	"time"
)

type User struct {
	ID int64

	Name string

	CreatedAt time.Time
	DeletedAt time.Time
	Deleted   bool
}

func (u *User) Emails() ([]Email, error) {
	var emails []Email
	rows, err := global.dao.GetEmailsForUser(context.Background(), u.ID)
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

func (u *User) Callsigns() ([]Callsign, error) {
	var callsigns []Callsign
	rows, err := global.dao.GetCallsignsForUser(context.Background(), u.ID)
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
	ErrUserNeedsCallsign = errors.New("user needs a callsign")
	ErrUserNeedsName     = errors.New("user needs a name")
)

func (m *User) Ready() error {
	var errs []error

	{
		rows, err := global.dao.GetCallsignsForUser(context.Background(), m.ID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			errs = append(errs, ErrUserNeedsCallsign)
		}
	}

	{
		if m.Name == "" {
			errs = append(errs, ErrUserNeedsName)
		}
	}

	return errors.Join(errs...)
}
