package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type Dashboard struct {
	views views.Dashboard
}

func init() {
	global.handlers = append(global.handlers, Dashboard{})
}

func (h Dashboard) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)

	r.Get(named.Route("dashboard-index", "/"), h.Index)
}

func (h Dashboard) Index(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)

	if services.Session.IsAuthenticated(ctx) {
		account := services.Session.MustGetAccount(ctx)
		v := views.Dashboard{
			Account: account,
			Ready:   account.Ready(),
		}
		v.Authenticated().Render(ctx, w)
	} else {
		h.views.Anonymous().Render(ctx, w)
	}
}
