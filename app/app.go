package app

import (
	"embed"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"

	"github.com/ryanfaerman/netctl/app/graph"
	"github.com/ryanfaerman/netctl/app/resolver"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
)

//go:generate go run github.com/99designs/gqlgen generate
//go:generate sqlc generate

var Logger = log.Default()

//go:embed sql/migrations/*.sql
var migrations embed.FS

func Register() {

	rslvr, err := resolver.New(Logger, migrations)
	if err != nil {
		Logger.Error("cannot create resolver", "err", err)
	}
	web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {
		srv := handler.NewDefaultServer(
			graph.NewExecutableSchema(graph.Config{Resolvers: rslvr}),
		)

		srv.AddTransport(&transport.Websocket{
			Upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		})

		srv.AddTransport(&transport.SSE{})
		e.Payload.Routes().Handle(
			named.Route("graph", "/query"),
			srv,
		)
		if config.Flag.Get("graph.playground", true) {
			e.Payload.Routes().Handle(
				named.Route("graph_playground", config.Get("graph.playground.route", "/-/playground")),
				playground.Handler("GraphQL playground", "/query"),
			)
		}
	})
}
