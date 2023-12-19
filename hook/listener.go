package hook

// Listener is a function that can be registered with a Hook and can respond to
// events.
type Listener[T any] func(Event[T])
