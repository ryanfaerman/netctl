package middleware

import (
	"net/http"

	"github.com/ryanfaerman/netctl/web"

	"github.com/ryanfaerman/netctl/internal/views"
)

func HTMXOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "" {
			web.LogWith(r.Context(), "hx", "true")
			views.Errors{}.Unsupported().Render(r.Context(), w)
			return
		}
		next.ServeHTTP(w, r)
	})
}
