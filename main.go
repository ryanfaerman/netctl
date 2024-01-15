package main

import (
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/events"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
)

// type Event[K any] struct {
// 	At        time.Time
// 	SessionID int64
// 	Source    string
// 	Event     K
// }
//
// type Checkin struct {
// 	Callsign string
// 	Name     string
// }
//
// func (e Checkin) Apply(event Event[NetEvent], net *models.Net) {
// 	c := models.NetCheckin{
// 		Callsign: e.Callsign,
// 		At:       event.At,
// 		Kind:     models.NetCheckinKindRoutine,
// 	}
// 	net.Sessions[event.SessionID].Checkins = append(net.Sessions[event.SessionID].Checkins, c)
// }
//
// type AckCheckin struct {
// 	Callsign string
// }
//
// func (e AckCheckin) Apply(event Event[NetEvent], net *models.Net) {
// 	session := net.Sessions[event.SessionID]
// 	for i, checkin := range session.Checkins {
// 		if checkin.Callsign == e.Callsign {
// 			session.Checkins[i].Acked = true
// 		}
// 	}
// }
//
// type NetStarted struct{}
//
// func (e NetStarted) Apply(event Event[NetEvent], net *models.Net) {
// 	session := models.NetSession{
// 		ID:     event.SessionID,
// 		Status: models.NetStatusOpened,
// 	}
// 	net.Sessions[event.SessionID] = &session
// }
//
// type NetClosed struct{}
// type NetReopened struct{}
// type NetScheduled struct{}
//
// func (e NetScheduled) Apply(event Event[NetEvent], net *models.Net) {
// 	session := models.NetSession{
// 		ID:     event.SessionID,
// 		Status: models.NetStatusScheduled,
// 	}
// 	net.Sessions[event.SessionID] = &session
// }
//
// type NetEvent interface {
// 	Apply(Event[NetEvent], *models.Net)
// }

func main() {
	// history := []any{
	// 	NetStarted{},
	//
	// 	AcknowledgedCheckin{checkin{callsign: "W0TLM", name: "Ryan"}},
	// 	AcknowledgedCheckin{checkin{callsign: "W4BUG", name: "James"}},
	// 	NetClosed{},
	// 	NetReopened{},
	// 	NetCheckin{checkin{callsign: "ABC123", name: "Earl"}},
	// 	AcknowledgedCheckin{checkin{callsign: "ABC123", name: "Earl"}},
	// 	NetClosed{},
	// }

	history := models.EventStream[events.Net]{
		{At: time.Now(), StreamID: "abc", Originator: 1, Event: events.NetStarted{}},
		{At: time.Now().Add(1 * time.Minute), StreamID: "abc", Originator: 1, Event: events.NetCheckin{Callsign: "KQ4JXI", Name: "Ryan"}},
		{At: time.Now().Add(1 * time.Minute), StreamID: "abc", Originator: 1, Event: events.NetCheckin{Callsign: "W4BUG", Name: "James"}},
		{At: time.Now(), StreamID: "xyz", Originator: 1, Event: events.NetStarted{}},
		{At: time.Now().Add(2 * time.Minute), StreamID: "abc", Originator: 1, Event: events.NetAckCheckin{Callsign: "KQ4JXI"}},
	}
	spew.Dump(history)

	n := NewFromHistory(history)
	fmt.Println(n)

}

func NewFromHistory(events []models.Event[events.Net]) *models.Net {
	net := &models.Net{
		ID:       314,
		Name:     "Test Net",
		Sessions: make(map[string]*models.NetSession, 0),
	}
	for _, event := range events {
		services.Net.SaveEvent(event.StreamID, event.Event)
		event.Event.Apply(event, net)
	}

	return net
}
