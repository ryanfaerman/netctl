package handlers

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"

	. "github.com/ryanfaerman/netctl/internal/models/finders"
	"modernc.org/sqlite"
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

	r.Get(named.Route("verify-email", "/settings/{slug}/emails/-/verify"), h.VerifyEmail)
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
	case "geolocation":
		settings = account.Settings.LocationSettings
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
	case "emails":
		settings = &models.Email{}
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
		email        models.Email
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
	case "geolocation":
		err = global.form.Decode(&settings.LocationSettings, r.Form)
		viewSettings = settings.LocationSettings
	case "emails":
		err = global.form.Decode(&email, r.Form)
		viewSettings = &email

	}

	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	settingsErrs := map[string]string{}

	switch namespace {
	case "emails":
		if err := services.Validation.Apply(email); err != nil {
			if errs, ok := err.(services.ValidationError); ok {
				spew.Dump(errs)

				for field, e := range errs {
					switch field {
					case "Email.Address":
						settingsErrs["email"] = e
					}
				}

				v.ShowWithErrors(namespace, viewSettings, settingsErrs).Render(ctx, w)
				return
			}

			ErrorHandler(err)(w, r)
			return
		}
	default:
		// Account.SaveSettings validates settings as whole and has trouble
		// when the input is a zero value. Until the service can handle zero values better,
		// we need to validate this here.
		if err := services.Validation.Apply(viewSettings); err != nil {
			if errs, ok := err.(services.ValidationError); ok {
				for field, e := range errs {
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
					case "LocationSettings.Latitude":
						settingsErrs["latitude"] = e
					case "LocationSettings.Longitude":
						settingsErrs["longitude"] = e
					case "LocationSettings.TimeOffset":
						settingsErrs["timeOffset"] = e

					}
				}

				v.ShowWithErrors(namespace, viewSettings, settingsErrs).Render(ctx, w)
				return
			}
			ErrorHandler(err)(w, r)
			return

		}
	}

	switch namespace {
	case "emails":
		fmt.Println("SAVED EMAILS")
		if err := services.Account.AddEmail(ctx, account.ID, &email); err != nil {
			spew.Dump(err)
			if e, ok := err.(*sqlite.Error); ok {
				switch e.Code() {
				case 2067:
					settingsErrs["email"] = "This email is already in use."
					v.ShowWithErrors(namespace, viewSettings, settingsErrs).Render(ctx, w)
					return
				}
			}
			ErrorHandler(err)(w, r)
			return
		}
	default:
		if err := services.Account.SaveSettings(ctx, account.ID, &settings); err != nil {
			ErrorHandler(err)(w, r)
			return
		}
	}
	v.Show(namespace, viewSettings).Render(ctx, w)
}

func (h settings) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	slug := chi.URLParam(r, "slug")
	token := r.URL.Query().Get("token")
	account, err := FindOne[models.Account](ctx, BySlug(slug))
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	if err := services.Email.VerifyEmailAddition(ctx, account, token); err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	if account.Kind != models.AccountKindUser {
		http.Redirect(w, r, named.URLFor("delegated-settings", slug, "emails"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, named.URLFor("settings", "emails"), http.StatusSeeOther)
}
