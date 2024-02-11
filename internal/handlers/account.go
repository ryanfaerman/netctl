package handlers

import (
	"net/http"

	"github.com/go-chi/chi"

	validator "github.com/go-playground/validator/v10"
	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type account struct{}

func init() {
	global.handlers = append(global.handlers, account{})
}

func (h account) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)
	r.Group(func(r chi.Router) {
		r.Use(middleware.HTMXOnly)
		// TODO: add an authenticated only middleware
		//
		r.Post(named.Route("account-setup-apply", "/account/setup"), h.Setup)
	})

	r.Get(named.Route("account-profile", "/profile/{slug}"), h.Show)
	r.Get(named.Route("account-profile-self", "/profile"), h.Show)
}

func (h account) Setup(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	inputErrs := views.SetupAccountErrors{}
	input := views.SetupAccountInput{
		Name:     r.Form.Get("name"),
		Callsign: r.Form.Get("callsign"),
	}

	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		for field, e := range errs.Translate(trans) {
			switch field {
			case "SetupaccountInput.Name":
				inputErrs.Name = e
			case "SetupaccountInput.Callsign":
				inputErrs.Callsign = e
			}
		}
		ctx := services.CSRF.GetContext(r.Context(), r)

		(views.Dashboard{}).SetupAccountWithErrors(input, inputErrs).Render(ctx, w)
		return
	}

	account := services.Session.GetAccount(r.Context())
	if err := services.Account.Setup(r.Context(), account.ID, input.Name, input.Callsign); err != nil {

		global.log.Error("unable to setup account", "error", err)
		switch err {
		case services.ErrAccountSetupCallsignTaken:
			inputErrs.Callsign = "Callsign is already taken"
		case services.ErrAccountSetupInvalidCallsign:
			inputErrs.Callsign = "Callsign not found in FCC database"
		case services.ErrAccountSetupCallsignClub:
			inputErrs.Callsign = "Callsign must be for an individual"
		}
		ctx := services.CSRF.GetContext(r.Context(), r)

		(views.Dashboard{}).SetupAccountWithErrors(input, inputErrs).Render(ctx, w)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	// TODO: update the profile, kick back errors if there are any
}

func (h account) Show(w http.ResponseWriter, r *http.Request) {
	var (
		account *models.Account
		err     error
	)

	slug := chi.URLParam(r, "slug")

	user := services.Session.GetAccount(r.Context())
	if user.IsAnonymous() && slug == "" {
		ErrorHandler(services.ErrNotAuthorized)(w, r)
		return
	}

	if slug == "" && slug != user.Slug {
		http.Redirect(w, r, named.URLFor("account-profile", user.Callsign().Call), http.StatusSeeOther)
		return
	}

	if slug != user.Slug {
		account, err = services.Account.FindBySlug(r.Context(), slug)
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
	} else {
		account = user
	}

	if err := services.Authorization.Can(r.Context(), user, "view", account); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	ctx := services.CSRF.GetContext(r.Context(), r)

	account.About = services.Markdown.MustRenderString(account.About)

	v := views.Account{
		Account: account,
	}
	v.Profile().Render(ctx, w)
}
