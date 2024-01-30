package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/middleware"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
)

type Markdown struct{}

func init() {
	global.handlers = append(global.handlers, Markdown{})
}

func (h Markdown) Routes(r chi.Router) {
	r.Use(middleware.HTMXOnly)
	web.CSRFExempt("/-/tools/markdown-render/*")
	r.Post(named.Route("markdown-preview", "/-/tools/markdown-render/{name}"), h.Preview)
	r.Post(named.Route("markdown-editor", "/-/tools/markdown-editor/{name}"), h.Editor)
}

func (h Markdown) Preview(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// ctx := services.CSRF.GetContext(r.Context(), r)
	field := chi.URLParam(r, "name")
	if field == "" {
		ErrorHandler(errors.New("no field name provided"))(w, r)
		return
	}

	attrs, err := views.DecodeInputAttrs(r.Form.Get(fmt.Sprintf("_%s-config", field)))
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	attrs.Value = r.Form.Get(field)
	attrs.MarkdownModePreview = true
	if attrs.Value != "" {
		attrs.MarkdownPreviewBody = services.Markdown.MustRenderString(attrs.Value)
	}

	views.InputTextArea(field, attrs).Render(r.Context(), w)
}

func (h Markdown) Editor(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// ctx := services.CSRF.GetContext(r.Context(), r)
	field := chi.URLParam(r, "name")
	if field == "" {
		ErrorHandler(errors.New("no field name provided"))(w, r)
		return
	}

	attrs, err := views.DecodeInputAttrs(r.Form.Get(fmt.Sprintf("_%s-config", field)))
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	attrs.Value = r.Form.Get(field)
	attrs.MarkdownModePreview = false

	views.InputTextArea(field, attrs).Render(r.Context(), w)
}
