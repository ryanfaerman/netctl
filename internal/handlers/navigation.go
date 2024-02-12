package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type Navigation struct{}

func init() {
	global.handlers = append(global.handlers, Navigation{})
}

func (h Navigation) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)
	r.Use(middleware.HTMXOnly)
	r.Get(named.Route("slide-over-show", "/-/slide-over/show/{side}"), h.SlideOverShow)
	r.Get(named.Route("slide-over-hide", "/-/slide-over/hide"), h.SlideOverHide)

	r.Get(named.Route("modal-show", "/-/modal/show/{name}"), h.ModalShow)
	r.Get(named.Route("modal-hide", "/-/modal/hide"), h.ModalHide)
}

func (h Navigation) SlideOverShow(w http.ResponseWriter, r *http.Request) {
	side := chi.URLParam(r, "side")
	if side != "left" && side != "right" {
		side = "right"
	}
	if side == "right" {
		views.RightNav(r.Context()).Show().Render(r.Context(), w)
	} else {
		views.LeftNav(r.Context()).Show().Render(r.Context(), w)
	}
}

func (h Navigation) SlideOverHide(w http.ResponseWriter, r *http.Request) {
	views.SlideOverTarget().Render(r.Context(), w)
}

func (h Navigation) ModalShow(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	v := views.Modal{
		Name: name,
	}

	switch name {
	default:
		h.ModalHide(w, r)
		return
	case "settings-context-switcher":
		currentUser := services.Session.GetAccount(r.Context())
		accounts := []*models.Account{currentUser}
		delegates, err := currentUser.Delegated(r.Context())
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		for _, d := range delegates {
			target := d.Target(r.Context())
			accounts = append(accounts, target)
		}
		v.SettingsContextSwitcher(accounts).Render(r.Context(), w)
	}
}

func (h Navigation) ModalHide(w http.ResponseWriter, r *http.Request) {
	views.ModalOverlay(false).Render(r.Context(), w)
}
