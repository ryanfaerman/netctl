package frontend

import (
	"context"
	"embed"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
	scs "github.com/alexedwards/scs/v2"
	branca "github.com/essentialkaos/branca/v2"
	"github.com/go-chi/chi"
	"github.com/justinas/nosurf"
	sse "github.com/r3labs/sse/v2"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/web/named"

	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/web"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

type Frontend struct {
	html *HTML

	events  *sse.Server
	brc     branca.Branca
	session *scs.SessionManager
}

//go:embed static/*
var staticFS embed.FS

func Register() {
	f := &Frontend{
		html: &HTML{
			title:       config.Get("frontend.title", "Netctl"),
			description: config.Get("frontend.description", "Netctl"),
			author:      config.Get("frontend.author", "Ryan Faerman"),
		},

		events:  sse.New(),
		session: scs.New(),
	}

	if br, err := branca.NewBranca([]byte(config.Get("random_key"))); err != nil {
		panic(err)
	} else {
		f.brc = br
	}

	f.session.Cookie.Name = config.Get("session.name", "_session")
	f.session.Cookie.Path = config.Get("session.path", "/")

	f.html.session = f.session

	web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {
		e.Payload.Routes().Group(func(r chi.Router) {
			r.Use(f.session.LoadAndSave)

			r.Get(named.Route("root", "/"), f.IndexHandler)
			r.Get(named.Route("net-index", "/net"), f.Render(f.html.NetIndex()))

			r.Get(named.Route("user-login", "/session/new"), f.IndexHandler)

			r.Group(func(r chi.Router) {
				r.Use(f.htmxOnly)

				r.Post(named.Route("session-create", "/session/create"), f.SessionCreateHandler)
			})
			r.Get(named.Route("session-verify", "/session/verify"), f.SessionVerifyHandler)
			r.Get(named.Route("session-destroy", "/session/destroy"), f.SessionDestroyHandler)
		})

		static, _ := fs.Sub(staticFS, "static")
		e.Payload.Routes().Handle(
			"/static/*",
			http.StripPrefix("/static/", statigz.FileServer(static.(fs.ReadDirFS), brotli.AddEncoding)),
		)

	})

	web.HookServerStop.Register(func(e hook.Event[web.ServerStopPayload]) {
		f.events.Close()
	})

}

func (f *Frontend) htmxOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "" {
			web.LogWith(r.Context(), "hx", "true")
			f.html.Unsupported().Render(r.Context(), w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (f *Frontend) Render(c templ.Component) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		web.LogWith(r.Context(), "nosurf_token", nosurf.Token(r))
		ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))
		c.Render(ctx, w)
	}
}
