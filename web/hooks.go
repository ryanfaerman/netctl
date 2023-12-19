package web

import (
	"context"
	"net"
	"net/http"

	"github.com/ryanfaerman/netctl/hook"
)

type ServerRoute struct {
	Route   string
	Handler http.HandlerFunc
}

type ServerStartPayload struct {
	Server   Server
	Listener net.Listener
}

type ServerStopPayload struct {
	Context context.Context
}

var (
	HookServer       = hook.New[Server]("server")
	HookServerStart  = hook.New[ServerStartPayload]("server.start")
	HookServerStop   = hook.New[ServerStopPayload]("server.stop")
	HookServerRoutes = hook.New[Router]("server.routes")
)
