package models

import "time"

//go:generate stringer -type=NetStatus -trimprefix=NetStatus
type NetStatus int

const (
	NetStatusUnknown NetStatus = iota
	NetStatusScheduled
	NetStatusOpened
	NetStatusClosed
)

type Period struct {
	OpenedAt time.Time
	ClosedAt time.Time
}

func (p Period) IsOpen() bool   { return p.ClosedAt.IsZero() }
func (p Period) IsClosed() bool { return !p.ClosedAt.IsZero() }

func (p Period) Duration() time.Duration {
	if p.IsClosed() {
		return p.ClosedAt.Sub(p.OpenedAt)
	}
	return time.Now().Sub(p.OpenedAt)
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
