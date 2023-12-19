package web

import (
	"github.com/chmike/securecookie"
	"github.com/ryanfaerman/netctl/config"
)

func init() {
	config.Flag.Define("web.debug", false)

	config.Define("random.key", string(securecookie.MustGenerateRandomKey()))
	config.Define("session.name", "_netctl_session")
}
