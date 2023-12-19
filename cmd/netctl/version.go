package main

import (
	"fmt"
	"net/http"

	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/version"
	"github.com/spf13/cobra"
)

var (
	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Display the version info",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.Version)
		},
	}
)

func init() {
	root.AddCommand(
		cmdVersion,
	)

	web.HookServerRoutes.Register(func(e hook.Event[web.Router]) {
		e.Payload.Routes().Get("/.well-known/version", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, version.Version)
		})
	})
}
