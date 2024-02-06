package handlers

import (
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

	spew.Dump(r.Form)
	if err := global.form.Decode(&group, r.Form); err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	spew.Dump(group)

	if err := services.Membership.Create(ctx, services.Session.GetAccount(ctx), &group); err != nil {
		// TODO: handle this more gracefully
		ErrorHandler(err)(w, r)
		return
	}

	switch kind {
	case models.AccountKindClub:
		w.Header().Set("HX-Redirect", named.URLFor("settings", "clubs"))
	case models.AccountKindOrganization:
		w.Header().Set("HX-Redirect", named.URLFor("settings", "organizations"))
	}
	return

	v := views.Membership{
		Kind: kind,
	}
	v.Create().Render(ctx, w)
}

func (h Membership) CheckSlug(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	kind := models.ParseAccountKind(chi.URLParam(r, "kind"))
	name := r.Form.Get("name")
	spew.Dump(r.Form)

	v := views.Membership{
		Kind: kind,
	}
	slug := r.Form.Get("slug")
	if slug == "" {
		slug = services.Slugger.Generate(r.Context(), name)
	}
	var err error
	slug, err = services.Slugger.ValidateUniqueForAccount(r.Context(), slug)
	if err != nil {
		v.SlugField(slug, "That organization ID is already in use").Render(r.Context(), w)
		return
	}

	v.SlugField(slug, "").Render(r.Context(), w)
}
