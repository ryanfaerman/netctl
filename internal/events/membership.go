package events

import "time"

func init() {
	register[MembershipRequested]("membership.requested")
	register[MembershipAccepted]("membership.accepted")
	register[MembershipDenied]("membership.denied")
	register[MembershipRevoked]("membership.revoked")
}

type (
	MembershipRequested struct {
		Message string `json:"message"`
	}

	MembershipAccepted struct {
		Message string `json:"message"`
	}

	MembershipDenied struct {
		Until   time.Time `json:"until"`
		Message string    `json:"message"`
	}

	MembershipRevoked struct {
		Message string `json:"message"`
	}

	MembershipEnded struct {
		Message string `json:"message"`
	}
)
