package events

import (
	"errors"
	"time"
)

func init() {
	register[NetCheckinHeard]("net.checkin_heard")
	register[NetCheckinVerified]("net.checkin_verified")
	register[NetCheckinAcked]("net.checkin_acked")
	register[NetCheckinCorrected]("net.checkin_corrected")
	register[NetCheckinRevoked]("net.checkin_revoked")

	register[NetSessionScheduled]("net.session_scheduled")
	register[NetSessionOpened]("net.session_opened")
	register[NetSessionClosed]("net.session_closed")
}

var (
	ErrorTypeNone         = errors.New("not an error")
	ErrorTypeNotFound     = errors.New("not found")
	ErrorTypeClubClass    = errors.New("club class")
	ErrorTypeLookupFailed = errors.New("lookup failed")
)

type (
	// NetCheckinHeard is a struct that contains the information about a checkin
	// as it was heard by the net control operator.
	NetCheckinHeard struct {
		ID       string `json:"id"` // a random identifier, ideally a ULID
		Callsign string `json:"callsign"`
		Name     string `json:"name"`
		Location string `json:"location"`
		Kind     string `json:"kind"` // e.g. models.EventCheckinKindRoutine
		Traffic  int    `json:"traffic"`
	}

	// NetCheckinVerified is a struct that contains the information about a checkin
	// as it was verified against a licensing authority.
	NetCheckinVerified struct {
		ID         string `json:"id"` // should match the ID from NetCheckinHeard
		Callsign   string `json:"callsign"`
		Name       string `json:"name"`
		Location   string `json:"location"`
		ErrorType  string `json:"error_type"`  // e.g. hamdb.ErrNotFound, ErrClub
		CallsignID int64  `json:"callsign_id"` // reference a record in the database
	}

	// NetCheckinAcked is a struct that contains the information about a checkin
	// as it was acknowledged by the net control operator.
	NetCheckinAcked struct {
		ID string `json:"id"` // should match the ID from NetCheckinHeard
	}

	// NetCheckinCorrected is a struct that contains the information about a checkin
	// as it was corrected by the net control operator.
	NetCheckinCorrected struct {
		ID       string `json:"id"` // should match the ID from NetCheckinHeard
		Callsign string `json:"callsign"`
		Name     string `json:"name"`
		Location string `json:"location"`
		Kind     string `json:"kind"`
		Traffic  int    `json:"traffic"`
	}

	// NetCheckinRevoked is a struct that contains the information about a checkin
	// as it was revoked by the net control operator.
	NetCheckinRevoked struct {
		ID string `json:"id"` // should match the ID from NetCheckinHeard
	}

	// NetSessionScheduled occurs when a net session is scheduled. Often occurs
	// when it is first created.
	NetSessionScheduled struct {
		At time.Time `json:"at"`
	}

	// NetSessionOpened occurs when a net session is opened.
	NetSessionOpened struct{}

	// NetSessionClosed occurs when a net session is closed.
	NetSessionClosed struct{}
)
