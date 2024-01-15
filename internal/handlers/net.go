package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type Net struct{}

func init() {
	global.handlers = append(global.handlers, Net{})
}

func (h Net) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)
	r.Get(named.Route("net-index", "/nets"), h.Index)
	r.Get(named.Route("net-show", "/nets/{id}"), h.Show)

	r.Post(named.Route("net-session-new", "/nets/{id}/new"), h.CreateSession)
	r.Get(named.Route("net-session-show", "/nets/{id}/{session_id}"), h.SessionShow)

	// r.Get(named.Route("net-checkin", "/nets/{id}/checkin"), h.Checkin)
	r.Post(named.Route("net-session-checkin", "/nets/{id}/{session_id}/checkin"), h.Checkin)
	// r.Post(named.Route("net-ack-checkin", "/nets/{id}/ack-checkin"), h.AckCheckin)
}

func (h Net) Index(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	nets, err := services.Net.All(ctx)
	if err != nil {
		global.log.Error("unable to get nets", "error", err)
	}
	v := views.Net{
		Nets: nets,
	}
	v.List().Render(ctx, w)
}

func (h Net) Show(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}
	net, err := services.Net.Get(ctx, id)
	if err != nil {
		global.log.Error("unable to get net", "error", err)
		panic("at the disco")
		return
	}
	v := views.Net{
		Net: net,
	}
	v.Show().Render(ctx, w)
	// // TODO: validate stream exists
	// v.SingleNetSession(stream).Render(ctx, w)
}

func (h Net) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}
	net, err := services.Net.Get(ctx, id)
	if err != nil {
		global.log.Error("unable to get net", "error", err)
		panic("at the disco")
		return
	}
	session, err := net.AddSession(ctx)
	if err != nil {
		global.log.Error("unable to add session", "error", err)
		panic("at the disco")
		return
	}
	http.Redirect(w, r, named.URLFor("net-session-show", strconv.FormatInt(id, 10), session.ID), http.StatusFound)
}

func (h Net) SessionShow(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}
	net, err := services.Net.GetReplayed(ctx, id)
	if err != nil {
		global.log.Error("unable to get net", "error", err)
		panic("at the disco")
		return
	}
	sessionID := chi.URLParam(r, "session_id")
	session, ok := net.Sessions[sessionID]
	if !ok {
		global.log.Error("unable to get session", "error", err)
		panic("at the disco")
		return
	}

	// if err := services.Event.Create(ctx, sessionID, events.NetCheckin{
	// 	Callsign: "W0RLI",
	// 	Name:     "Fred",
	// }); err != nil {
	// 	global.log.Error("unable to create event", "error", err)
	// }

	v := views.Net{
		Net:     net,
		Session: session,
	}
	v.SingleNetSession(sessionID).Render(ctx, w)
}

func (h Net) Checkin(w http.ResponseWriter, r *http.Request) {
	// ctx := services.CSRF.GetContext(r.Context(), r)
	// id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	// if err != nil {
	// 	panic(err)
	// }
	// net, err := services.Net.Get(ctx, id)
	// if err != nil {
	// 	global.log.Error("unable to get net", "error", err)
	// 	panic("at the disco")
	// 	return
	// }
	// sessionID := chi.URLParam(r, "session_id")
	// session, ok := net.Sessions[sessionID]
	// if !ok {
	// 	global.log.Error("unable to get session", "error", err)
	// 	panic("at the disco")
	// 	return
	// }
	// if err := services.Net.Checkin(ctx, id, sessionID); err != nil {
	// 	global.log.Error("unable to checkin", "error", err)
	// 	panic("at the disco")
	// 	return
	// }
	// http.Redirect(w, r, named.URLFor("net-session-show", strconv.FormatInt(id, 10), session.ID), http.StatusFound)
}
