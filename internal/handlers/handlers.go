package handlers

import (
	"net/http"

	scs "github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	sse "github.com/r3labs/sse/v2"
	"github.com/ryanfaerman/netctl/config"

	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/web"
)

var global struct {
	events   *sse.Server
	session  *scs.SessionManager
	handlers []routable
}

type routable interface {
	Routes(r chi.Router)
}

type routableFunc func(r chi.Router)

func (r routableFunc) Routes(router chi.Router) { r(router) }

func registerRoutableFunc(r routableFunc) {
	global.handlers = append(global.handlers, r)
}

func Register() {
	global.events = sse.New()
	global.session = scs.New()

	global.session.Cookie.Name = config.Get("session.name", "_session")
	global.session.Cookie.Path = config.Get("session.path", "/")

	web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {
		for _, h := range global.handlers {
			e.Payload.Routes().Group(h.Routes)
		}

	})

	web.HookServerStop.Register(func(e hook.Event[web.ServerStopPayload]) {
		global.events.Close()
	})

}

func htmxOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "" {
			web.LogWith(r.Context(), "hx", "true")
			// f.html.Unsupported().Render(r.Context(), w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// func (f *Frontend) Render(c templ.Component) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		web.LogWith(r.Context(), "nosurf_token", nosurf.Token(r))
// 		ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))
// 		c.Render(ctx, w)
// 	}
// }
