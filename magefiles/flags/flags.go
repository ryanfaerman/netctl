package flags

import (
	"fmt"
	"strings"

	"github.com/ryanfaerman/netctl/magefiles/module"
)

// LDFlags wraps the generation of LDFLags meant to be passed to the compiler
type LDFlags map[string]string

// Build the LDFlags for the given package
func (ldf LDFlags) Build(packageName string) string {
	var b strings.Builder

	for k, v := range ldf {
		if strings.HasPrefix(k, "./") {
			fmt.Fprintf(&b, `-X "%s/%s=%s" `, packageName, strings.TrimPrefix(k, "./"), v)
		} else {
			fmt.Fprintf(&b, `-X "%s=%s" `, k, v)
		}
	}

	return b.String()
}

// String returns the LDFlags for the detected ModulePath.
func (ldf LDFlags) String() string {
	return ldf.Build(module.Path())
}
