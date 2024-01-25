package handlers

import (
	"database/sql"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi"
	sse "github.com/r3labs/sse/v2"

	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/web"
)

var global = struct {
	events   *sse.Server
	handlers []routable
	log      *log.Logger
	db       *sql.DB
}{
	log: log.With("pkg", "handlers"),
}

func ogger(l *log.Logger) { global.log = l.With("pkg", "handlers") }

type routable interface {
	Routes(r chi.Router)
}

type routableFunc func(r chi.Router)

func (r routableFunc) Routes(router chi.Router) { r(router) }

func registerRoutableFunc(r routableFunc) {
	global.handlers = append(global.handlers, r)
}

var setupOnce sync.Once

func Setup(logger *log.Logger, db *sql.DB) error {
	var err error

	setupOnce.Do(func() {
		global.log = logger.With("pkg", "handlers")

		global.log.Debug("running setup tasks")
		global.db = db

		global.events = sse.New()
		global.events.AutoReplay = false
		global.events.AutoStream = true
		services.Event.Server = global.events

		web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {
			for _, h := range global.handlers {
				e.Payload.Routes().Group(h.Routes)
			}
		})

		web.HookServerStop.Register(func(e hook.Event[web.ServerStopPayload]) {
			global.events.Close()
		})
	})

	return err
}

// func (f *Frontend) Render(c templ.Component) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		web.LogWith(r.Context(), "nosurf_token", nosurf.Token(r))
// 		ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))
// 		c.Render(ctx, w)
// 	}
// }
