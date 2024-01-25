package handlers

import (
	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/web/named"
)

type Event struct{}

func init() {
	global.handlers = append(global.handlers, Event{})
}

func (h Event) Routes(r chi.Router) {
	r.HandleFunc(named.Route("sse-source", "/-/events"), global.events.ServeHTTP)
}
