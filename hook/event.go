package hook

import "context"

// Event is the type that is passed to listeners when a hook is dispatched.
type Event[T any] struct {
	// Payload is our type being passed along to the listener
	Payload T

	// Error is any error that should be bubbled up to the dispatcher
	Error error

	// Context is passed along from the dispatcher
	Context context.Context

	// Hook is the hook that dispatched the event. Used to disambiguate events
	Hook *Hook[T]

	// ID is the internal ID of the event
	id uint64
}

func newEvent[T any](msg T, hook *Hook[T]) Event[T] {
	return Event[T]{
		Payload: msg,
		Hook:    hook,
		Context: context.Background(),
	}
}

// Unregister the event from the hook. This is useful if you no longer want to
// receive events.
func (e *Event[T]) Unregister() {
	e.Hook.unregister(e.id)
}

// ID get the internal ID of the event
func (e *Event[T]) ID() uint64 {
	return e.id
}
