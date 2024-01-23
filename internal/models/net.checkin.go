package models

import "time"

type Hearable struct {
	AsHeard    string
	AsLicensed string
}

type NetCheckin struct {
	ID string
	At time.Time

	Callsign Hearable
	Location Hearable
	Name     Hearable

	Acked    bool
	Verified bool
	Valid    error
	Kind     NetCheckinKind
	Traffic  int
}
