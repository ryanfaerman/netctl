package web

import (
	"github.com/a-h/templ"
	"net/http"
)

func RenderComponent(c templ.Component) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Render(r.Context(), w)
	}
}
