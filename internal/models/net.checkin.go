package models

import "time"

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
