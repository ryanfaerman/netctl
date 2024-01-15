package handlers

import (
	"net/http"

	"github.com/go-chi/chi"

	validator "github.com/go-playground/validator/v10"
	"github.com/ryanfaerman/netctl/internal/middleware"
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

	// callsign, err := hamdb.Lookup(r.Context(), input.Callsign)
	// if err != nil { ,
	// 	inputErrs.Callsign = "Could not validate callsign"
	// 	if err != hamdb.ErrNotFound {
	// 		global.log.Error("hamdb lookup failed", "error", err)
	// 	}
	//
	// 	ctx := services.CSRF.GetContext(r.Context(), r)
	//
	// 	(views.Dashboard{}).SetupaccountWithErrors(input, inputErrs).Render(ctx, w)
	// 	return
	// }
	// spew.Dump(callsign)

	account, err := services.Session.GetAccount(r.Context())
	if err != nil {
		global.log.Error("unable to get account from session", "error", err)
		//views.Errors{}.Internal().Render(r.Context(), w)
		return
	}
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
