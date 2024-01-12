package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/web/named"
)

type User struct{}

func init() {
	global.handlers = append(global.handlers, User{})
}

func (h User) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.HTMXOnly)
		// TODO: add an authenticated only middleware
		//
		r.Post(named.Route("user-update-profile", "/user/update-profile"), h.UpdateProfile)
	})
}

func (h User) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// TODO: update the profile, kick back errors if there are any
}
