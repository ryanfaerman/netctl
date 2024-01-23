package events

import (
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(NetCheckinHeard{})
	gob.Register(NetCheckinVerified{})
	gob.Register(NetCheckinAcked{})
	gob.Register(NetCheckinCorrected{})
	gob.Register(NetCheckinRevoked{})

	gob.Register(NetSessionScheduled{})
	gob.Register(NetSessionOpened{})
	gob.Register(NetSessionClosed{})
}

type (
	// NetCheckinHeard is a struct that contains the information about a checkin
	// as it was heard by the net control operator.
	NetCheckinHeard struct {
		ID string // a random identifier, ideally a ULID

		Callsign string
		Name     string
		Location string
		Kind     string // e.g. models.EventCheckinKindRoutine
		Traffic  int
	}

	// NetCheckinVerified is a struct that contains the information about a checkin
	// as it was verified against a licensing authority.
	NetCheckinVerified struct {
		ID string // should match the ID from NetCheckinHeard

		Callsign string
		Name     string

		ErrorType string // e.g. hamdb.ErrNotFound, ErrClub

		CallsignID int64 // reference a record in the database
	}

	// NetCheckinAcked is a struct that contains the information about a checkin
	// as it was acknowledged by the net control operator.
	NetCheckinAcked struct {
		ID string // should match the ID from NetCheckinHeard
	}

	// NetCheckinCorrected is a struct that contains the information about a checkin
	// as it was corrected by the net control operator.
	NetCheckinCorrected struct {
		ID string // should match the ID from NetCheckinHeard

		Callsign string
		Name     string
		Location string
		Kind     string
		Traffic  int
	}

	// NetCheckinRevoked is a struct that contains the information about a checkin
	// as it was revoked by the net control operator.
	NetCheckinRevoked struct {
		ID string // should match the ID from NetCheckinHeard
	}

	// NetSessionScheduled occurs when a net session is scheduled. Often occurs
	// when it is first created.
	NetSessionScheduled struct {
		At time.Time
	}

	// NetSessionOpened occurs when a net session is opened.
	NetSessionOpened struct{}

	// NetSessionClosed occurs when a net session is closed.
	NetSessionClosed struct{}
)
