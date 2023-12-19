package config

import (
	"github.com/ryanfaerman/netctl/hook"
)

type Definition struct {
	c *config
}

func (d Definition) Flag(uri string, fallback ...bool) {
	d.c.DefineFlag(uri, fallback...)
}

func (d Definition) Config(uri string, fallback ...string) {
	d.c.Define(uri, fallback...)
}

var (
	Hook = hook.New[Definition]("config.definitions")
)
