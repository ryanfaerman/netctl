package handlers

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	sse "github.com/r3labs/sse/v2"

	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
)

type Net struct{}

func init() {
	global.handlers = append(global.handlers, Net{})
}

func (h Net) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)

	r.Get(named.Route("net-index", "/nets"), h.Index)
	r.Get(named.Route("net-new", "/nets/new"), h.New)
	r.Get(named.Route("net-show", "/net/{net_id}"), h.Show)

	r.Get(named.Route("net-session-show", "/net/session/{session_id}"), h.SessionShow)
	r.Post(named.Route("net-session-new", "/net/{net_id}/new"), h.CreateSession)

	r.Get(named.Route("get-checkin", "/net/{net_id}/{session_id}/{checkin_id}"), h.CheckinShow)

	r.Group(func(r chi.Router) {
		r.Use(middleware.HTMXOnly)

		r.Post(named.Route("net-create", "/nets/create"), h.Create)
		r.Post(named.Route("net-session-checkin", "/net/{net_id}/{session_id}/checkin"), h.Checkin)

		web.CSRFExempt("/net/*/ack/*")
		r.Post(named.Route("checkin-ack", "/net/{session_id}/ack/{checkin_id}"), h.AckCheckin)
	})
}

func (h Net) Index(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	nets, err := services.Net.All(ctx)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Net{
		Nets: nets,
	}
	v.List().Render(ctx, w)
}

func (h Net) Show(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	net_id := chi.URLParam(r, "net_id")
	net, err := services.Net.GetByStreamID(ctx, net_id)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Net{
		Net: net,
	}
	v.Show().Render(ctx, w)
}

func (h Net) New(w http.ResponseWriter, r *http.Request) {
	v := views.Net{}

	ctx := services.CSRF.GetContext(r.Context(), r)
	v.Create().Render(ctx, w)
}

func (h Net) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	v := views.Net{}

	ctx := services.CSRF.GetContext(r.Context(), r)

	inputErrs := views.CreateNetFormErrors{}
	input := views.CreateNetFormInput{
		Name: r.Form.Get("name"),
	}

	m, err := services.Net.Create(r.Context(), models.NewNet(input.Name))
	if err != nil {
		if errs, ok := err.(services.ValidationError); ok {
			for field, e := range errs {
				switch field {
				case "models.Net.Name":
					inputErrs.Name = e
				}
			}
		} else {
			inputErrs.Name = "Unable to create the net. Please try again later."
			global.log.Error("Net creation failed", "error", err)
		}

		v.CreateFormWithErrors(input, inputErrs).Render(ctx, w)
		return
	}

	v.Net = m
	v.Show().Render(ctx, w)
}

// CreateSession creates a new session for a net and
// redirects to the session page
func (h Net) CreateSession(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	net_id := chi.URLParam(r, "net_id")

	session, err := services.Net.CreateSession(ctx, net_id)
	if err != nil {
		ref := ErrorHandler(err)(w, r)
		global.log.Error("unable to create session", "error", err, "ref", ref)
		return
	}

	http.Redirect(w, r, named.URLFor("net-session-show", session.ID), http.StatusFound)
}

func (h Net) SessionShow(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	sessionID := chi.URLParam(r, "session_id")

	net, err := services.Net.GetNetFromSession(ctx, sessionID)
	if err != nil {
		ref := ErrorHandler(err)(w, r)
		global.log.Error("unable to get net", "error", err, "ref", ref)
		return
	}

	// eventStream, err := net.Events(ctx, sessionID)
	// if err != nil {
	// 	ErrorHandler(err)(w, r)
	// 	return
	// }

	v := views.Net{
		Net:     net,
		Session: net.Sessions[sessionID],
		// Stream:  eventStream,
	}

	v.SingleNetSession(sessionID).Render(ctx, w)
}

func (h Net) Checkin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	net_id := chi.URLParam(r, "net_id")
	net, err := services.Net.GetByStreamID(r.Context(), net_id)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	sessionID := chi.URLParam(r, "session_id")
	session, ok := net.Sessions[sessionID]
	if !ok {
		ErrorHandler(err)(w, r)
		return
	}
	v := views.Net{
		Net:     net,
		Session: session,
	}
	ctx := services.CSRF.GetContext(r.Context(), r)

	inputErrs := views.CheckinFormErrors{}
	input := views.CheckinFormInput{
		Callsign: strings.ToUpper(strings.TrimSpace(r.Form.Get("call-sign"))),
		Name:     strings.TrimSpace(r.Form.Get("name")),
		Traffic:  strings.TrimSpace(r.Form.Get("traffic")),
	}

	m, err := services.Net.Checkin(r.Context(), sessionID, &models.NetCheckin{
		Callsign: models.Hearable{AsHeard: input.Callsign},
		Name:     models.Hearable{AsHeard: input.Name},
		Kind:     models.ParseNetCheckinKind(input.Traffic),
	})
	if err != nil {
		if errs, ok := err.(services.ValidationError); ok {
			for field, e := range errs {
				switch field {
				case "NetCheckin.Callsign.AsHeard":
					inputErrs.Callsign = e
				case "NetCheckin.Name.AsHeard":
					inputErrs.Name = e
				case "NetCheckin.Kind":
					inputErrs.Traffic = e
				}
			}
			v.CheckinFormWithErrors(input, inputErrs).Render(ctx, w)
			return
		}

		if err == services.ErrCheckinExists {
			v.CheckinForm().Render(ctx, w)

			services.Event.Server.Publish(sessionID, &sse.Event{
				Event: []byte(m.ID),
				Data:  []byte("found"),
			})
			return
		} else {
			global.log.Error("Checkin failed", "error", err)

			inputErrs.Name = "Server Error: Unable to perform checkin"
			v.CheckinFormWithErrors(input, inputErrs).Render(ctx, w)
			return
		}
	}

	v.CheckinForm().Render(ctx, w)

	var b bytes.Buffer
	v.CheckinRow(*m, true).Render(ctx, &b)
	w.Write(b.Bytes())
	services.Event.Server.Publish(sessionID, &sse.Event{
		Data: b.Bytes(),
	})
}

func (h Net) CheckinShow(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	net_id := chi.URLParam(r, "net_id")
	net, err := services.Net.GetByStreamID(ctx, net_id)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	sessionID := chi.URLParam(r, "session_id")
	session, ok := net.Sessions[sessionID]
	if !ok {
		ErrorHandler(err)(w, r)
		return
	}

	net.Replay(ctx, sessionID)
	v := views.Net{
		Net:     net,
		Session: session,
	}

	for _, checkin := range session.Checkins {
		if checkin.ID == chi.URLParam(r, "checkin_id") {
			v.CheckinRow(checkin).Render(ctx, w)
			return
		}
	}
}

func (h Net) AckCheckin(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	stream := chi.URLParam(r, "session_id")
	id := chi.URLParam(r, "checkin_id")
	err := services.Net.AckCheckin(ctx, stream, id)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}
}
