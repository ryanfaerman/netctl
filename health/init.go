package health

import "github.com/ryanfaerman/netctl/config"

func init() {
	config.Define("health.route", "/.well-known/ruok")
	config.Flag.Define("health.enabled", true)
}
