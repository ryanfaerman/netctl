package frontend

import (
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

func (f *Frontend) IndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := contextWithCSRFToken(r.Context(), r)

	for _, k := range f.session.Keys(ctx) {
		fmt.Println("key", k)
		spew.Dump(f.session.Get(ctx, k))
	}

	if f.IsAuthenticated(ctx) {
		f.html.AuthenticatedDashboard().Render(ctx, w)
	} else {
		f.html.AnonymousDashboard().Render(ctx, w)
	}
}
