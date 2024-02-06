package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"runtime"

	"github.com/oklog/ulid/v2"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web"
)

func ErrorHandler(err error) func(w http.ResponseWriter, r *http.Request) string {
	ref := ulid.Make().String()
	_, file, line, _ := runtime.Caller(1)
	return func(w http.ResponseWriter, r *http.Request) string {
		if err == nil {
			return ref
		}

		web.LogWith(r.Context(), "ref", ref, "error", err, "file", file, "line", line)

		v := views.Errors{
			Error:     err,
			Reference: ref,
		}

		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			v.NotFound().Render(r.Context(), w)
			return ref
		}

		if errors.Is(err, services.ErrNotAuthorized) {
			w.WriteHeader(http.StatusForbidden)
			v.Unauthorized().Render(r.Context(), w)
			return ref
		}

		w.WriteHeader(http.StatusInternalServerError)
		v.General().Render(r.Context(), w)
		global.log.Error("unknown error", "ref", ref, "error", err)

		return ref
	}
}
