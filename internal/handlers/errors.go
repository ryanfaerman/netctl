package handlers

import (
	"database/sql"
	"net/http"

	"github.com/oklog/ulid/v2"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web"
)

func ErrorHandler(err error) func(w http.ResponseWriter, r *http.Request) string {
	ref := ulid.Make().String()
	return func(w http.ResponseWriter, r *http.Request) string {
		if err == nil {
			return ref
		}

		web.LogWith(r.Context(), "ref", ref, "error", err)

		v := views.Errors{
			Error:     err,
			Reference: ref,
		}

		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			v.NotFound().Render(r.Context(), w)
			return ref
		}

		w.WriteHeader(http.StatusInternalServerError)
		v.General().Render(r.Context(), w)
		global.log.Error("unknown error", "ref", ref, "error", err)

		return ref
	}
}
