package models

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/dao"
)

type MockAggregate struct {
	status  string
	message string
}

type MockStatusChanged struct{}  // status change
type MockMessageChanged struct { // message change
	Message string
}

func init() {
	gob.Register(MockStatusChanged{})
	gob.Register(MockMessageChanged{})
}

func TestEvent(t *testing.T) {

	accountID, err := global.dao.CreateAccountAndReturnId(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	netID, err := global.dao.CreateNetAndReturnId(context.Background(), "test net")
	if err != nil {
		t.Fatal(err)
	}

	{
		// create some events
		events := []Event[any]{
			{StreamID: "abc", Originator: accountID, Event: MockStatusChanged{}},
			{StreamID: "abc", Originator: accountID, Event: MockMessageChanged{"hello there"}},
		}
		for _, event := range events {
			var b bytes.Buffer
			var p any
			p = &event.Event
			if err := gob.NewEncoder(&b).Encode(p); err != nil {
				t.Fatal(err)
			}

			if err := global.dao.CreateNetEvent(context.Background(), dao.CreateNetEventParams{
				NetID:     netID,
				SessionID: event.StreamID,
				AccountID: event.Originator,
				EventType: fmt.Sprintf("%T", event.Event),
				EventData: b.Bytes(),
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	events, err := global.dao.GetNetEvents(context.Background(), netID)
	if err != nil {
		t.Fatal(err)
	}

	for _, raw := range events {

		decoder := gob.NewDecoder(bytes.NewReader(raw.EventData))
		var p any
		if err := decoder.Decode(&p); err != nil {
			t.Fatal(err)
		}
		event := Event[any]{
			ID:         raw.ID,
			At:         raw.Created,
			StreamID:   raw.SessionID,
			Originator: raw.AccountID,
			Name:       raw.EventType,
			Event:      p,
		}

		switch e := event.Event.(type) {
		case MockStatusChanged:
			fmt.Println("status changed")
		case MockMessageChanged:
			fmt.Println("message changed", e.Message)

		}

		spew.Dump(p)
	}

}
func getTypeByName(typeName string) reflect.Type {
	// Iterate over all registered types in the program.
	fmt.Println("typeName:", typeName)
	switch typeName {
	case "models.MockStatusChanged":
		return reflect.TypeOf(MockStatusChanged{})
	case "models.MockMessageChanged":
		return reflect.TypeOf(MockMessageChanged{})
	}

	return nil
}
