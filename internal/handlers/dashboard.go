package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type Dashboard struct {
	views    views.Dashboard
	services struct {
		session services.Session
	}
}

func init() {
	global.handlers = append(global.handlers, Dashboard{})
}

func (h Dashboard) Routes(r chi.Router) {
	r.Use(h.services.session.Middleware)

	r.Get(named.Route("dashboard-index", "/"), h.Index)
}

func (h Dashboard) Index(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)

	if h.services.session.IsAuthenticated(ctx) {
		h.views.Authenticated().Render(ctx, w)
	} else {
		h.views.Anonymous().Render(ctx, w)
	}
}
