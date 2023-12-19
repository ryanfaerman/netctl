package hook

import (
	"sync"
	"testing"
)

const (
	hookName      = "test.hook"
	listenerCount = 3
)

// message holds the data being dispatched via the test hooks
type message struct {
	id int
}

// listener facilitate the testing of hook listeners
type listener struct {
	counter int
	msg     *message
	hook    *Hook[*message]
	wg      sync.WaitGroup
	t       *testing.T
}

// newListener creates and initializes a new listener with a hook and message
func newListener(t *testing.T) (*listener, *Hook[*message], *message) {
	msg := &message{id: 123}
	h := New[*message](hookName)
	l := &listener{
		t:    t,
		msg:  msg,
		hook: h,
	}

	for i := 0; i < listenerCount; i++ {
		h.Register(l.Callback)
	}

	l.wg.Add(listenerCount)

	if listenerCount != h.ListenerCount() {
		t.Fail()
	}

	return l, h, msg
}

// Callback is the callback method for the test hooks that counts executions, confirms the event data, and
// handles waitgroups for concurrency
func (l *listener) Callback(event Event[*message]) {
	l.counter++

	if l.msg != event.Payload {
		l.t.Fail()
	}

	if l.hook != event.Hook {
		l.t.Fail()
	}

	if hookName != event.Hook.Name() {
		l.t.Fail()
	}

	l.wg.Done()
}
