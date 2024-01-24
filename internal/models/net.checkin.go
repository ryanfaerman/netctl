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

	Acked    bool  // when the checkin is acked by net control
	Verified bool  // if verification has been performed
	Valid    error // any verification errors
	Kind     NetCheckinKind
	Traffic  int
}
