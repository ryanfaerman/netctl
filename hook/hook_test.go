package hook

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/charmbracelet/log"
)

type Message struct {
	Greeting string
}

func TestHookUnregister(t *testing.T) {

	h := New[Message]("test.hook.unregister")

	called := 0

	for i := 0; i < 3; i++ {
		i := i
		h.Register(func(e Event[Message]) {
			called++

			if i == 1 {
				e.Unregister()
			}

			if e.Payload.Greeting != "Hello There!" {
				t.Errorf("payload should be 'Hello There!', got %s", e.Payload.Greeting)
			}
		})
	}

	if h.ListenerCount() != 3 {
		t.Error("listener count should be 3")
	}

	h.Dispatch(context.Background(), Message{Greeting: "Hello There!"})

	if h.ListenerCount() != 2 {
		t.Errorf("listener count should be 2, got %d", h.ListenerCount())
	}

	if called != 3 {
		t.Errorf("called should be 3, got %d", called)
	}

}

func TestHookDispatch(t *testing.T) {
	l, h, msg := newListener(t)

	h.Dispatch(context.Background(), msg)
	l.wg.Wait()

	if listenerCount != l.counter {
		t.Fail()
	}
}

func TestConcurrentUnregister(t *testing.T) {

	if testing.Verbose() {
		Logger.SetLevel(log.DebugLevel)
	}

	ch := make(chan struct{}, 10)
	h := New[Message]("test.hook.concurrent_unregister")

	called := 0
	unregistered := 0
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		h.Register(func(e Event[Message]) {
			wg.Add(1)
			defer wg.Done()
			select {
			case ch <- struct{}{}:
				called++
			default:
				unregistered++
				e.Unregister()
				return
			}
		})
	}

	h.Dispatch(context.Background(), Message{Greeting: "Hello There!"})
	wg.Wait()
	fmt.Println("called", called, "unregistered", unregistered)
	// close(ch)

	if called != 10 {
		t.Errorf("called should be 10, got %d", called)
	}
	if unregistered != 90 {
		t.Errorf("unregistered should be 90, got %d", unregistered)
	}

	// drain the channel and reset our counters
	for i := 0; i < called; i++ {
		<-ch
	}
	called = 0
	unregistered = 0

	h.Dispatch(context.Background(), Message{Greeting: "Hello There!"})
	wg.Wait()
	if called != 10 {
		t.Errorf("called should be 10, got %d", called)
	}
	if unregistered != 0 {
		t.Errorf("unregistered should be 0, got %d", unregistered)
	}

}
