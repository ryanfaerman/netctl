package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ryanfaerman/netctl/web"
)

var (
	Logger = log.Default()
)

func handle(w http.ResponseWriter, r *http.Request) {
	sh := Check{
		Points: make(map[string]error),
	}

	Logger.Debug("triggering health.check")

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	dispatchErr := Hook.Dispatch(ctx, &sh)

	ruok := dispatchErr == nil
	output := map[string]interface{}{}

	fails := map[string]string{}
	if !ruok {
		fails["ruok"] = dispatchErr.Error()
	}

	for name, err := range sh.Points {
		ok := err == nil
		output[name] = ok

		if !ok {
			ruok = false
			fails[name] = err.Error()
		}
	}

	output["ruok"] = ruok
	output["errors"] = fails

	switch web.GetContentType(r) {
	case web.ContentTypeJSON:
		web.Respond(r).JSON(w, http.StatusOK, output)
	default:
		web.Respond(r).Text(w, http.StatusOK, fmt.Sprintf("%v\n", ruok))
	}
}
