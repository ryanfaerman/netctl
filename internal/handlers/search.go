package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/internal/views"
	"github.com/ryanfaerman/netctl/web/named"
)

type Search struct{}

func init() {
	global.handlers = append(global.handlers, Search{})
}

func (h Search) Routes(r chi.Router) {
	r.Use(services.Session.Middleware)
	r.Get(named.Route("find-local", "/find/local/{kind}"), h.Search)
}

func (h Search) Search(w http.ResponseWriter, r *http.Request) {
	distanceStr := r.URL.Query().Get("distance")
	unit := r.URL.Query().Get("unit")

	kind := models.ParseAccountKind(chi.URLParam(r, "kind"))

	if distanceStr == "" {
		distanceStr = "5"
	}
	if unit == "" {
		unit = "miles"
	}

	v := views.Search{
		Distance: distanceStr,
		Unit:     unit,
		Kind:     kind,
	}

	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	if strings.ToLower(unit) == "miles" {
		distance = distance * 1.609344
	}

	account := services.Session.GetAccount(r.Context())
	if account.IsAnonymous() {
		ErrorHandler(fmt.Errorf("you must be logged in to use this feature"))(w, r)
		return
	}
	lat, lon, err := services.Account.Geolocation(r.Context(), account)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}

	input := services.CallsignsWithinRangeParams{
		Latitude:  lat,
		Longitude: lon,
		Distance:  distance,
		Kind:      int(kind),
	}

	results, err := services.Search.CallsignsWithinRange(r.Context(), input)
	if err != nil {
		ErrorHandler(err)(w, r)
		return
	}
	v.Found = results
	if r.Header.Get("HX-Request") == "true" {
		v.ResultList().Render(r.Context(), w)
		return
	}
	v.Results().Render(r.Context(), w)
}
