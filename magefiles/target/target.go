package target

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/ryanfaerman/netctl/magefiles/module"
)

type Target struct {
	GOOS   string
	GOARCH string
}

func New(goos, goarch string) Target {
	return Target{
		GOOS:   goos,
		GOARCH: goarch,
	}
}

// Local returns a target for the local GOOS and GOARCH
func Local() Target {
	return New(runtime.GOOS, runtime.GOARCH)
}

func (t Target) Name() string {
	n := filepath.Base(module.Path())
	name := fmt.Sprintf(
		"%s-%s-%s",
		n,
		t.GOOS,
		t.GOARCH,
	)

	if t.GOOS == "windows" {
		name += ".exe"
	}

	return name
}

func (t Target) Env() map[string]string {
	return map[string]string{
		"GOOS":   t.GOOS,
		"GOARCH": t.GOARCH,
	}
}
