package events

import "fmt"

type subscriber struct {
	connection chan any
}

type Bus struct {
	subscribers map[string][]subscriber
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]subscriber),
	}
}

func (b *Bus) Subscribe(event any) chan any {
	c := make(chan any)
	s := subscriber{connection: c}
	b.subscribers[fmt.Sprintf("%T", event)] = append(b.subscribers[fmt.Sprintf("%T", event)], s)
	return c
}

func (b *Bus) Publish(event any) {
	for _, s := range b.subscribers[fmt.Sprintf("%T", event)] {
		s.connection <- event
	}
}
