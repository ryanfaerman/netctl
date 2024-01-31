package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type Navigation struct{}

func init() {
	global.handlers = append(global.handlers, Navigation{})
}

func (h Navigation) Routes(r chi.Router) {
	r.Use(middleware.HTMXOnly)
	r.Get(named.Route("slide-over-show", "/-/slide-over/show"), h.SlideOverShow)
	r.Get(named.Route("slide-over-hide", "/-/slide-over/hide"), h.SlideOverHide)
}

func (h Navigation) SlideOverShow(w http.ResponseWriter, r *http.Request) {
	views.SlideOver().Render(r.Context(), w)
}

func (h Navigation) SlideOverHide(w http.ResponseWriter, r *http.Request) {
	views.SlideOverTarget().Render(r.Context(), w)
}