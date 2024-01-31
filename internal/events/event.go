package events

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
)

// A registrant is a type registered with the event system
// and available for decoding.
type registrant struct {
	new      func() any
	friendly string // a friendly name for the type, e.g. "user.session_created"
	name     string // the name of the type, e.g. "events.UserCreated"
}

// Registry holds the registered types.
var registry = make(map[string]registrant)

// Register a type with the event system. We collect the name of the event,
// a friendly type for user/consumer reference, and a function that returns a pointer
// to a new instance of the type. We also register the type with gob, should
// a consumer wish to encode/decode the event to/from a binary format.
//
// The friendly name shouldn't be used for anything other than display. Internal
// references should use the name field.
func register[K any](friendly string) {
	gob.Register(*new(K))

	name := fmt.Sprintf("%T", *new(K))
	registry[name] = registrant{
		name: name,
		new: func() any {
			return new(K)
		},
		friendly: friendly,
	}
}

// Decode an event from JSON. The kind is the name of the event, and the data is the
// JSON payload. This is useful for decoding events from the database or queue.
func Decode(kind string, data []byte) (any, error) {
	k := registry[kind].new()
	return k, json.Unmarshal(data, k)
}

// FriendlyNameFor returns the friendly name for the event type, as provided
// during registration.
func FriendlyNameFor(kind string) string {
	return registry[kind].friendly
}

// NameFor returns the name of the event type, e.g. "events.UserCreated".
func NameFor(kind string) string {
	return registry[kind].name
}
