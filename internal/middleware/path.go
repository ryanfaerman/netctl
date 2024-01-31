package middleware

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/web/named"
)

func Path(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := chi.RouteContext(r.Context())
		spew.Dump(c.RoutePattern(), c.URLParams)
		fmt.Println("lookup", named.Lookup(c.RoutePattern()))
		next.ServeHTTP(w, r)
	})
}
