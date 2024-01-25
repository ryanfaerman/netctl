package models

import (
	"strings"
	"time"
)

//go:generate stringer -type=NetStatus -trimprefix=NetStatus
type NetStatus int

const (
	NetStatusUnknown NetStatus = iota
	NetStatusScheduled
	NetStatusOpened
	NetStatusClosed
)

type Period struct {
	OpenedAt  time.Time
	ClosedAt  time.Time
	Scheduled bool
}

func (p Period) IsOpen() bool      { return p.ClosedAt.IsZero() }
func (p Period) IsClosed() bool    { return !p.ClosedAt.IsZero() }
func (p Period) IsScheduled() bool { return p.Scheduled }

func (p Period) Duration() time.Duration {
	if p.IsClosed() {
		return p.ClosedAt.Sub(p.OpenedAt)
	}
	return time.Since(p.OpenedAt)
}

type Periods []Period

func (p Periods) Duration() time.Duration {
	var d time.Duration
	for _, period := range p {
		d += period.Duration()
	}
	return d
}

type NetSession struct {
	CreatedAt time.Time
	ID        string
	Periods   Periods

	Checkins []NetCheckin
}

func (m NetSession) FindCheckinByCallsign(call string) *NetCheckin {
	for _, checkin := range m.Checkins {
		if strings.ToUpper(checkin.Callsign.AsHeard) == strings.ToUpper(call) {
			return &checkin
		}
	}
	return nil
}
