package handlers

import (
	"net/http"
	"strings"

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

	r.Get(named.Route("account-profile", "/profile/{callsign}"), h.Show)
	r.Get(named.Route("account-edit", "/profile/{callsign}/edit"), h.Edit)
	r.Post(named.Route("account-edit-save", "/profile/{callsign}/edit/-/save"), h.Update)
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

	account, err := services.Session.GetAccount(r.Context())
	if err != nil {
		global.log.Error("unable to get account from session", "error", err)
		// views.Errors{}.Internal().Render(r.Context(), w)
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

func (h account) Show(w http.ResponseWriter, r *http.Request) {
	callsign := chi.URLParam(r, "callsign")
	ctx := services.CSRF.GetContext(r.Context(), r)

	a, err := services.Account.FindByCallsign(ctx, callsign)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Account{
		Account: a,
	}
	v.Profile().Render(ctx, w)
}

func (h account) Edit(w http.ResponseWriter, r *http.Request) {
	callsign := chi.URLParam(r, "callsign")
	ctx := services.CSRF.GetContext(r.Context(), r)
	a, err := services.Account.FindByCallsign(ctx, callsign)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	v := views.Account{
		Account: a,
	}
	v.Edit().Render(ctx, w)
}

func (h account) Update(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	callsign := chi.URLParam(r, "callsign")
	ctx := services.CSRF.GetContext(r.Context(), r)
	a, err := services.Account.FindByCallsign(ctx, callsign)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	v := views.Account{
		Account: a,
	}

	inputErrs := views.AccountEditFormErrors{}
	input := views.AccountEditFormInput{
		Name: strings.TrimSpace(r.Form.Get("name")),
	}
	a.Name = input.Name
	err = services.Account.Update(r.Context(), a)
	if err != nil {
		if errs, ok := err.(services.ValidationError); ok {
			for field, e := range errs {
				switch field {
				case "Account.Name":
					inputErrs.Name = e
				}
			}
			v.EditFormWithErrors(input, inputErrs).Render(ctx, w)
			return
		}
		ErrorHandler(err)(w, r)
		return
	}
	// TODO: return this from the account.update method
	v.Account = a

	v.EditForm().Render(ctx, w)
}
