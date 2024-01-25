package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi"
	sse "github.com/r3labs/sse/v2"

	validator "github.com/go-playground/validator/v10"
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
	r.Post(named.Route("net-create", "/nets/create"), h.Create)
	r.Get(named.Route("net-show", "/net/{id}"), h.Show)

	r.Post(named.Route("net-session-new", "/net/{id}/new"), h.CreateSession)
	r.Get(named.Route("net-session-show", "/net/{id}/{session_id}"), h.SessionShow)

	// r.Get(named.Route("net-checkin", "/nets/{id}/checkin"), h.Checkin)
	r.Post(named.Route("net-session-checkin", "/net/{id}/{session_id}/checkin"), h.Checkin)
	r.Get(named.Route("get-checkin", "/net/{id}/{session_id}/{checkin_id}"), h.CheckinShow)

	r.Group(func(r chi.Router) {
		r.Use(middleware.HTMXOnly)
		web.CSRFExempt("/net/*/ack/*")
		r.Post(named.Route("checkin-ack", "/net/{session_id}/ack/{checkin_id}"), h.AckCheckin)
	})
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

func (h Net) New(w http.ResponseWriter, r *http.Request) {
	v := views.Net{}

	ctx := services.CSRF.GetContext(r.Context(), r)
	v.Create().Render(ctx, w)
}

func (h Net) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	v := views.Net{}

	inputErrs := views.CreateNetFormErrors{}
	input := views.CreateNetFormInput{
		Name: r.Form.Get("name"),
	}
	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		for field, e := range errs.Translate(trans) {
			switch field {
			case "CreateNetFormInput.Name":
				inputErrs.Name = e
			}
		}

		ctx := services.CSRF.GetContext(r.Context(), r)
		v.CreateFormWithErrors(input, inputErrs).Render(ctx, w)
		return
	}

	net, err := services.Net.Create(r.Context(), input.Name)
	if err != nil {
		global.log.Error("unable to create net", "error", err)
		panic("at the disco")
	}
	w.Header().Set("HX-Location", named.URLFor("net-show", strconv.FormatInt(net.ID, 10)))
	http.Redirect(w, r, named.URLFor("net-show", strconv.FormatInt(net.ID, 10)), http.StatusFound)
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
	net, err := services.Net.Get(ctx, id)
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
	net.Replay(ctx, sessionID)
	eventStream, err := net.Events(ctx, sessionID)
	if err != nil {
		global.log.Error("unable to get event stream", "error", err)
		panic("at the disco")
	}
	v := views.Net{
		Net:     net,
		Session: session,
		Stream:  eventStream,
	}

	if !global.events.StreamExists(sessionID) {
		global.events.CreateStream(sessionID)
	}

	v.SingleNetSession(sessionID).Render(ctx, w)
}

func (h Net) Checkin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}
	net, err := services.Net.Get(r.Context(), id)
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
	v := views.Net{
		Net:     net,
		Session: session,
	}

	inputErrs := views.CheckinFormErrors{}
	input := views.CheckinFormInput{
		Callsign: strings.ToUpper(strings.TrimSpace(r.Form.Get("call-sign"))),
		Name:     strings.TrimSpace(r.Form.Get("name")),
		Traffic:  strings.TrimSpace(r.Form.Get("traffic")),
	}
	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		for field, e := range errs.Translate(trans) {
			switch field {
			case "CheckinFormInput.Name":
				inputErrs.Name = e
			case "CheckinFormInput.Callsign":
				inputErrs.Callsign = e
			case "CheckinFormInput.Traffic":
				inputErrs.Traffic = e
			}
		}

		ctx := services.CSRF.GetContext(r.Context(), r)
		v.CheckinFormWithErrors(input, inputErrs).Render(ctx, w)
		return
	}

	checkin := models.NetCheckin{
		Callsign: models.Hearable{AsHeard: input.Callsign},
		Name:     models.Hearable{AsHeard: input.Name},
		Kind:     models.ParseNetCheckinKind(input.Traffic),
	}

	checkinID, err := services.Net.Checkin(r.Context(), sessionID, &checkin)
	if err != nil {
		global.log.Error("unable to checkin", "error", err)
		panic("at the disco")
		return
	}
	checkin.ID = checkinID

	net.Replay(r.Context(), sessionID)
	v.Session = net.Sessions[sessionID]

	ctx := services.CSRF.GetContext(r.Context(), r)
	v.CheckinForm().Render(ctx, w)
	// v.CheckinRow(checkin).Render(ctx, w)
	// for _, checkin := range session.Checkins {
	// 	if checkin.ID == checkinID {
	// 		v.CheckinRow(checkin).Render(ctx, w)
	// 		break
	// 	}
	// }

	found := v.Session.FindCheckinByCallsign(checkin.Callsign.AsHeard)
	fmt.Println("found")
	spew.Dump(found)
	fmt.Println("checkin")
	spew.Dump(checkin)

	var b bytes.Buffer

	se := &sse.Event{}
	if found.ID != checkin.ID {
		se.Event = []byte(found.ID)
		v.CheckinRow(*found).Render(ctx, &b)
		se.Data = b.Bytes()
		se.Data = []byte("found")
	} else {
		v.CheckinRow(*found, true).Render(ctx, &b)
		se.Data = b.Bytes()

	}

	global.events.Publish(sessionID, se)

	// http.Redirect(w, r, named.URLFor("net-session-show", strconv.FormatInt(id, 10), session.ID), http.StatusFound)
}

func (h Net) CheckinShow(w http.ResponseWriter, r *http.Request) {
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
	sessionID := chi.URLParam(r, "session_id")
	session, ok := net.Sessions[sessionID]
	if !ok {
		global.log.Error("unable to get session", "error", err)
		panic("at the disco")
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
		global.log.Error("unable to ack checkin", "error", err)
		panic("at the disco")
		return
	}
}
