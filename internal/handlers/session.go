package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	validator "github.com/go-playground/validator/v10"
	"github.com/ryanfaerman/netctl/web/named"

	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/internal/views"

	"github.com/ryanfaerman/netctl/internal/services"
)

const (
	SessionKeyAuthenticated = "authenticated"
	SessionKeyFlash         = "flash"
)

type Session struct {
	view views.Session
}

func init() {
	global.handlers = append(global.handlers, Session{})
}

func (h Session) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)

	r.Get(named.Route("user-login", "/session/new"), h.Create)

	r.Group(func(r chi.Router) {
		r.Use(middleware.HTMXOnly)

		r.Post(named.Route("session-create", "/session/create"), h.Create)
	})

	r.Get(named.Route("session-verify", "/session/verify"), h.Verify)
	r.Get(named.Route("session-destroy", "/session/destroy"), h.Destroy)
}

func (h Session) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	inputErrs := views.SessionCreateErrors{}
	input := views.SessionCreateInput{
		Email: r.Form.Get("email"),
	}

	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		for field, e := range errs.Translate(trans) {
			switch field {
			case "SessionCreateInput.Email":
				inputErrs.Email = e
			}
		}

		ctx := services.CSRF.GetContext(r.Context(), r)

		h.view.LoginWithErrors(input, inputErrs).Render(ctx, w)
		return
	}

	if err := services.Session.SendEmailVerification(r.Context(), input.Email); err != nil {
		views.Errors{}.GeneralError(err).Render(r.Context(), w)
	} else {
		h.view.Created().Render(r.Context(), w)
	}

}

func (h Session) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if err := services.Session.Verify(r.Context(), token); err != nil {
		switch err {
		case services.ErrTokenInvalid,
			services.ErrTokenExpired,
			services.ErrTokenMismatch,
			services.ErrTokenDecode:
			h.view.VerificationFailed(err).Render(r.Context(), w)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h Session) Destroy(w http.ResponseWriter, r *http.Request) {
	services.Session.Destroy(r.Context())
	http.Redirect(w, r, "/", http.StatusFound)
}
