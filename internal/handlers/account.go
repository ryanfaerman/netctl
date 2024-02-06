package handlers

import (
	"net/http"
	"strings"

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

	r.Get(named.Route("account-profile", "/profile/{callsign}"), h.Show)
	r.Get(named.Route("account-profile-self", "/profile"), h.Show)
	r.Post(named.Route("account-edit-save", "/profile/{callsign}/edit/-/save"), h.Update)

	// r.Get(named.Route("settings-profile", "/settings/profile"), h.Edit)
	// r.Get(named.Route("settings-privacy", "/settings/privacy"), h.SettingsPrivacy)
	// r.Get(named.Route("settings-avatar", "/settings/avatar"), h.SettingsAvatar)
	// r.Get(named.Route("settings-billing", "/settings/billing"), h.Edit)
	// r.Get(named.Route("settings-emails", "/settings/emails"), h.Edit)
	// r.Get(named.Route("settings-sessions", "/settings/sessions"), h.Edit)
	r.Get(named.Route("settings", "/settings/{namespace}"), h.Settings)
	r.Post(named.Route("settings-save", "/settings/{namespace}/-/save"), h.SettingsSave)
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

	callsign := chi.URLParam(r, "callsign")

	user := services.Session.GetAccount(r.Context())
	if user.IsAnonymous() && callsign == "" {
		ErrorHandler(services.ErrNotAuthorized)(w, r)
		return
	}

	if callsign == "" && callsign != user.Callsign().Call {
		http.Redirect(w, r, named.URLFor("account-profile", user.Callsign().Call), http.StatusSeeOther)
		return
	}

	if callsign != user.Callsign().Call {
		account, err = services.Account.FindByCallsign(r.Context(), callsign)
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
	} else {
		account = user
	}

	if err := services.Authorization.Can(user, "view", account); err != nil {
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

func (h account) Edit(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	a := services.Session.GetAccount(r.Context())

	if err := services.Authorization.Can(a, "edit", a); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Account{
		Account: a,
	}
	v.Settings("profile", nil).Render(ctx, w)
}

func (h account) Settings(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	namespace := chi.URLParam(r, "namespace")

	a := services.Session.GetAccount(ctx)

	if err := services.Authorization.Can(a, "edit", a); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Account{
		Account: a,
	}

	var settings any

	switch namespace {
	case "privacy":
		settings = a.Settings.PrivacySettings
	case "appearance":
		settings = a.Settings.AppearanceSettings
	case "clubs":
		clubs, err := a.Clubs(ctx)
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		v.Memberships = clubs
	case "organizations":
		orgs, err := a.Organizations(ctx)
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		v.Memberships = orgs
	}

	v.Settings(namespace, settings).Render(ctx, w)
}

func (h account) SettingsSave(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ctx := services.CSRF.GetContext(r.Context(), r)
	a := services.Session.GetAccount(ctx)
	if err := services.Authorization.Can(a, "edit", a); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	namespace := chi.URLParam(r, "namespace")

	var (
		settings     models.Settings
		viewSettings any
		err          error
	)

	v := views.Account{
		Account: a,
	}

	switch namespace {
	case "privacy":
		err = global.form.Decode(&settings.PrivacySettings, r.Form)
		viewSettings = settings.PrivacySettings
	case "appearance":
		err = global.form.Decode(&settings.AppearanceSettings, r.Form)
		viewSettings = settings.AppearanceSettings
	}

	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	if err := services.Account.SaveSettings(ctx, a.ID, &settings); err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	v.Settings(namespace, viewSettings).Render(ctx, w)
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
		Name:  strings.TrimSpace(r.Form.Get("name")),
		About: strings.TrimSpace(r.Form.Get("about")),
	}
	a.Name = input.Name
	a.About = input.About
	err = services.Account.Update(ctx, a)
	if err != nil {
		if errs, ok := err.(services.ValidationError); ok {
			for field, e := range errs {
				switch field {
				case "Account.Name":
					inputErrs.Name = e
				case "Account.About":
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
	services.Session.SetAccount(ctx, a)

	v.EditForm().Render(ctx, w)
}
