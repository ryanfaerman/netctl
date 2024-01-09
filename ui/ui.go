package ui

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/a-h/templ"
	scs "github.com/alexedwards/scs/v2"
	branca "github.com/essentialkaos/branca/v2"
	"github.com/go-chi/chi"
	validator "github.com/go-playground/validator/v10"
	"github.com/justinas/nosurf"
	sse "github.com/r3labs/sse/v2"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

type key int

const (
	ctxToken key = iota
)

func render(c templ.Component) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		web.LogWith(r.Context(), "nosurf_token", nosurf.Token(r))
		ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))

		c.Render(ctx, w)
	}
}

//go:embed static/*
var staticFS embed.FS

var records = []Record{}

var sseServer = sse.New()

var brc branca.Branca

var session *scs.SessionManager

func Register() {

	if br, err := branca.NewBranca([]byte(config.Get("random_key"))); err != nil {
		panic(err)
	} else {
		fmt.Println("Branca key generated", br)
		brc = br
	}

	session = scs.New()
	session.Cookie.Name = config.Get("session.name")
	session.Cookie.Path = "/"

	web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {

		e.Payload.Routes().Group(func(r chi.Router) {
			r.Use(session.LoadAndSave)

			r.Get(named.Route("net-session", "/net"), render(NetSession()))

			r.Get(named.Route("index", "/"), MagicLinkNewHandler)
			r.Post(named.Route("magic-link-create", "/magic_link/create"), MagicLinkCreateHandler)
			r.Get(named.Route("magic-link-sent", "/magic_link/sent"), render(MagicLinkSent()))
			r.Get(named.Route("magic-link-verify", "/magic_link/verify"), MagicLinkVerifyHandler)

			r.Get("/token", tokenHandler)

			r.Post(named.Route("receive-checkin", "/net/checkin"), checkinHandler)

			r.Route("/beta", func(r chi.Router) {
				r.Get(named.Route("beta-magic-link-new", "/"), render(MagicLinkIndex()))
			})
		})

		static, _ := fs.Sub(staticFS, "static")
		e.Payload.Routes().Handle(
			"/static/*",
			http.StripPrefix("/static/", statigz.FileServer(static.(fs.ReadDirFS), brotli.AddEncoding)),
		)

		sseServer.CreateStream("messages")
		e.Payload.Routes().Handle(named.Route("net-sse-src", "/net/sse"), sseServer)
	})
	//
	// web.HookServerStart.Register(func(e hook.Event[web.ServerStartPayload]) {
	// 	// session = e.Payload.Server.Session()
	// })

	web.HookServerStop.Register(func(e hook.Event[web.ServerStopPayload]) {
		sseServer.Close()
	})
}

func checkinHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	record := Record{
		Callsign: r.Form.Get("callsign"),
		Name:     r.Form.Get("name"),
	}

	checkinErrs := CheckinErrors{}

	if err := validate.Struct(record); err != nil {
		for _, err := range err.(validator.ValidationErrors) {

			switch err.Field() {
			case "Name":
				if err.Tag() == "required" {
					checkinErrs.Name = "Name is a required field"
				}
			case "Callsign":
				if err.Tag() == "required" {
					checkinErrs.Callsign = "Callsign is a required field"
				}
				if err.Tag() == "max" {
					checkinErrs.Callsign = "Callsign must be 6 characters or less"
				}
			}
		}
	} else {
		records = append(records, record)
	}

	var b bytes.Buffer
	netrecord(records).Render(r.Context(), &b)

	if sseServer.StreamExists("messages") {
		sseServer.Publish("messages", &sse.Event{
			Data: b.Bytes(),
		})
	}

	if r.Header.Get("HX-Request") != "" {
		web.LogWith(r.Context(), "hx", "true")
		ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))
		netrecord(records).Render(ctx, w)
		checkForm(checkinErrs).Render(ctx, w)
		return
	}

	fmt.Println("redirecting")

	http.Redirect(w, r, named.URLFor("index"), http.StatusSeeOther)

}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(nosurf.Token(r)))

	payload := "banana"
	token, err := brc.EncodeToString([]byte(payload))
	if err != nil {
		panic(err)
	}

	got_token := r.URL.Query().Get("token")
	if len(got_token) > 0 {
		fmt.Println("token", got_token)
		decoded, err := brc.DecodeString(got_token)
		if err != nil {
			panic(err)
		}

		fmt.Println("decoded payload", string(decoded.Payload()))
		fmt.Println("decoded timestamp", decoded.Timestamp())
		fmt.Println("decoded is expired", decoded.IsExpired(60))
	}

	web.LogWith(r.Context(), "branca.token", token)
}
