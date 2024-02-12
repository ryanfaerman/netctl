package handlers

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi"
	"github.com/go-playground/form"
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
	form     *form.Decoder
}{
	log:  log.With("pkg", "handlers"),
	form: form.NewDecoder(),
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
				e.Payload.Routes().NotFound(func(w http.ResponseWriter, r *http.Request) {
					spew.Dump(r)
				})
			}
		})

		web.HookServerStop.Register(func(e hook.Event[web.ServerStopPayload]) {
			global.events.Close()
		})
	})

	return err
}
