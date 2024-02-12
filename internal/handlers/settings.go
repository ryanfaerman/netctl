package handlers

import (
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"

	. "github.com/ryanfaerman/netctl/internal/models/finders"
)

type settings struct{}

func init() {
	global.handlers = append(global.handlers, settings{})
}

func (h settings) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)

	r.Get(named.Route("settings", "/settings/{namespace}"), h.Settings)
	r.Post(named.Route("settings-save", "/settings/{namespace}/-/save"), h.SettingsSave)

	r.Get(named.Route("delegated-settings", "/settings/{slug}/{namespace}"), h.Settings)
	r.Post(named.Route("delegated-settings-save", "/settings/{slug}/{namespace}/-/save"), h.SettingsSave)
}

func (h settings) Settings(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	namespace := chi.URLParam(r, "namespace")
	slug := chi.URLParam(r, "slug")

	// currentUser can come from cache, but we don't want
	// the cached version, in case a permission or other
	// account detail has changed. So we need to get the
	// account from the database every time.
	currentUser := services.Session.GetAccount(ctx)
	if currentUser.Cached {
		a, err := FindOne[models.Account](ctx, ByID(currentUser.ID))
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		currentUser = a
	}

	var account *models.Account
	if slug != "" {
		a, err := FindOne[models.Account](ctx, BySlug(slug))
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		account = a
	} else {
		account = currentUser
	}

	if err := services.Authorization.Can(ctx, currentUser, "edit", account); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Settings{
		Account:   account,
		Delegated: account.ID != currentUser.ID,
	}

	var settings any

	switch namespace {
	case "profile":
		settings = account.Settings.ProfileSettings
	case "privacy":
		settings = account.Settings.PrivacySettings
	case "appearance":
		settings = account.Settings.AppearanceSettings
	case "clubs":
		clubs, err := account.Clubs(ctx)
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		v.Memberships = clubs
	case "organizations":
		orgs, err := account.Organizations(ctx)
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		v.Memberships = orgs
	}

	v.Show(namespace, settings).Render(ctx, w)
}

func (h settings) SettingsSave(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ctx := services.CSRF.GetContext(r.Context(), r)
	namespace := chi.URLParam(r, "namespace")
	slug := chi.URLParam(r, "slug")

	// currentUser can come from cache, but we don't want
	// the cached version, in case a permission or other
	// account detail has changed. So we need to get the
	// account from the database every time.
	currentUser := services.Session.GetAccount(ctx)
	if currentUser.Cached {
		a, err := FindOne[models.Account](ctx, ByID(currentUser.ID))
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		currentUser = a
	}

	var account *models.Account
	if slug != "" {
		a, err := FindOne[models.Account](ctx, BySlug(slug))
		if err != nil {
			ErrorHandler(err)(w, r)
			return
		}
		account = a
	} else {
		account = currentUser
	}

	if err := services.Authorization.Can(r.Context(), currentUser, "edit", account); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	v := views.Settings{
		Account:   account,
		Delegated: account.ID != currentUser.ID,
	}

	var (
		settings     models.Settings
		viewSettings any
		err          error
	)

	switch namespace {
	case "profile":
		err = global.form.Decode(&settings.ProfileSettings, r.Form)
		viewSettings = settings.ProfileSettings
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

	settingsErrs := map[string]string{}

	// Account.SaveSettings validates settings as whole and has trouble
	// when the input is a zero value. Until the service can handle zero values better,
	// we need to validate this here.
	if err := services.Validation.Apply(viewSettings); err != nil {
		if errs, ok := err.(services.ValidationError); ok {
			for field, e := range errs {
				spew.Dump(errs)
				switch field {
				case "ProfileSettings.Name":
					settingsErrs["name"] = e
				case "ProfileSettings.About":
					settingsErrs["about"] = e
				case "PrivacySettings.Location":
					settingsErrs["location"] = e
				case "PrivacySettings.Visibility":
					settingsErrs["visibility"] = e
				case "AppearanceSettings.ActivityGraphs":
					settingsErrs["activityGraphs"] = e

				}
				v.ShowWithErrors(namespace, viewSettings, settingsErrs).Render(ctx, w)
				return
			}
		}
		ErrorHandler(err)(w, r)
		return

	}

	if err := services.Account.SaveSettings(ctx, account.ID, &settings); err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	v.Show(namespace, viewSettings).Render(ctx, w)
}
