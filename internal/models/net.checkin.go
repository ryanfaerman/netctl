package models

import (
	"errors"
	"time"
)

type Hearable struct {
	AsHeard    string `validate:"eq=|alphanum"`
	AsLicensed string `validate:"eq=|alphanum"`
}

type NetCheckin struct {
	ID string
	At time.Time

	Callsign Hearable `validate:"required"`
	Location Hearable
	Name     Hearable

	Acked    bool           // when the checkin is acked by net control
	Verified bool           // if verification has been performed
	Valid    error          // any verification errors
	Kind     NetCheckinKind `validate:"required"`
	Traffic  int            `validate:"gte=0"`
}

func (m *NetCheckin) Can(account *Account, action string) error {
	if account.IsAnonymous() {
		return errors.New("not authorized")
	}
	switch action {
	case "create":
		if account.ID > 1 {
			return errors.New("not authorized")
		}
	}
	return nil
}
