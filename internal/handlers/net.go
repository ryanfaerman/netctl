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
	// r.Get(named.Route("net-index", "/nets"), h.Index)
	r.Get(named.Route("net-show-session", "/nets/{id}/{stream}"), h.Show)
	// r.Get(named.Route("net-checkin", "/nets/{id}/checkin"), h.Checkin)
	// r.Post(named.Route("net-checkin", "/nets/{id}/checkin"), h.Checkin)
	// r.Post(named.Route("net-ack-checkin", "/nets/{id}/ack-checkin"), h.AckCheckin)
}

func (h Net) Show(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		panic(err)
	}
	// if err := services.Net.StartSession(ctx, id); err != nil {
	// 	global.log.Error("unable to start session", "error", err)
	// }
	// if err := services.Net.Checkin(ctx, id); err != nil {
	// 	global.log.Error("unable to start session", "error", err)
	// }
	net, err := services.Net.GetReplayed(ctx, id)
	if err != nil {
		global.log.Error("unable to get net", "error", err)
		panic("at the disco")
		return
	}
	eventStream, err := net.Events(ctx)
	if err != nil {
		global.log.Error("unable to get event stream", "error", err)
		return
	}
	stream := chi.URLParam(r, "stream")
	v := views.Net{
		Net:    net,
		Stream: eventStream.ForStream(stream),
	}
	// TODO: validate stream exists
	v.SingleNetSession(stream).Render(ctx, w)
}
