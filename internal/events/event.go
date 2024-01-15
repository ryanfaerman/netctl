package events

import (
	"encoding/gob"
)

func init() {
	gob.Register(NetStarted{})
	gob.Register(NetScheduled{})
	gob.Register(NetCheckin{})
	gob.Register(NetAckCheckin{})
}

type NetStarted struct{}

type NetScheduled struct{}

type NetCheckin struct {
	Callsign string
	Name     string
}

type NetAckCheckin struct {
	Callsign string
}
