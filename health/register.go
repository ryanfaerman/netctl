package health

import (
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
)

func Register() {
	web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {
		e.Payload.Routes().Get(named.Route("ruok", config.Get("health.route")), handle)
	})
}
