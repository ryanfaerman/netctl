package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
)

type Membership struct{}

func init() {
	global.handlers = append(global.handlers, Membership{})
}

func (h Membership) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)

	r.Get(named.Route("group-new", "/groups/{kind}/new"), h.New)
	r.Post(named.Route("group-create", "/groups/{kind}/create"), h.Create)

	web.CSRFExempt("/group/*/check-slug")
	r.Post(named.Route("group-check-slug", "/groups/{kind}/check-slug"), h.CheckSlug)
}

func (h Membership) New(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)

	kind := models.ParseAccountKind(chi.URLParam(r, "kind"))
	v := views.Membership{
		Kind: kind,
	}

	v.Create().Render(ctx, w)
}

func (h Membership) Create(w http.ResponseWriter, r *http.Request) {
	ctx := services.CSRF.GetContext(r.Context(), r)
	r.ParseForm()

	kind := models.ParseAccountKind(chi.URLParam(r, "kind"))

	group := models.Account{
		Settings: models.DefaultSettings,
		Kind:     kind,
	}

	if err := global.form.Decode(&group, r.Form); err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	if err := services.Membership.Create(ctx, services.Session.GetAccount(ctx), &group, r.Form.Get("email"), r.Form.Get("callsign")); err != nil {
		// TODO: handle this more gracefully
		spew.Dump(err)
		viewInput := views.MembershipCreateFormInput{
			Name:     group.Name,
			Email:    r.Form.Get("email"),
			Slug:     group.Slug,
			Callsign: r.Form.Get("callsign"),
		}
		viewErr := views.MembershipCreateFormError{}
		switch {
		case errors.Is(err, services.ErrAccountSetupInvalidCallsign):
			viewErr.Callsign = "Invalid callsign"
		case errors.Is(err, services.ErrClubRequiresCallsign):
			viewErr.Callsign = "Clubs require a callsign"
		case errors.Is(err, services.ErrAccountSetupCallsignTaken):
			viewErr.Callsign = "Callsign already in use"
		case errors.Is(err, services.ErrAccountSetupCallsignIndividual):
			viewErr.Callsign = "Callsign must be for a club, not an individual"
		case errors.Is(err, services.ErrCallsignCreationFailed):
			viewErr.Callsign = "Unable to save callsign, try again"
		default:
			if errs, ok := err.(services.ValidationError); ok {
				for field, e := range errs {
					switch field {
					case "Account.Settings.ProfileSettings.Name":
						viewErr.Name = e
					case "Account.Slug":
						viewErr.Slug = e
					}
				}
			} else {
				ErrorHandler(err)(w, r)
				return
			}
		}

		v := views.Membership{
			Kind: kind,
		}
		v.CreateFormWithError(viewInput, viewErr).Render(ctx, w)
		return
	}

	switch kind {
	case models.AccountKindClub:
		w.Header().Set("HX-Redirect", named.URLFor("settings", "clubs"))
	case models.AccountKindOrganization:
		w.Header().Set("HX-Redirect", named.URLFor("settings", "organizations"))
	}
}

func (h Membership) CheckSlug(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	kind := models.ParseAccountKind(chi.URLParam(r, "kind"))
	name := r.Form.Get("name")

	v := views.Membership{
		Kind: kind,
	}

	var slug string
	if r.URL.Query().Get("source") != "slug" {
		fmt.Println("WAT")
		slug = services.Slugger.Generate(r.Context(), name)
	} else {
		fmt.Println("GONZO")
		slug = r.Form.Get("slug")
	}
	var err error
	slug, err = services.Slugger.ValidateUniqueForAccount(r.Context(), slug)
	if err != nil {
		v.SlugField(views.MembershipCreateFormInput{Slug: slug}, views.MembershipCreateFormError{Slug: "That organization ID is already in use"}).Render(r.Context(), w)
		return
	}

	v.SlugField(views.MembershipCreateFormInput{Slug: slug}, views.MembershipCreateFormError{}).Render(r.Context(), w)
}
