package config

import (
	"github.com/ryanfaerman/netctl/hook"
)

// Define registers a new config option, with an optional default value.
func Define(uri string, d ...string) {
	Hook.Register(func(e hook.Event[Definition]) {
		e.Payload.Config(uri, d...)
	})
}

// Wait until the config is loaded.
func Wait() {
	c.WaitForLoad()
}

// IsConfig checks if the config uri is a config.
func IsConfig(uri string) bool {
	return c.IsConfig(uri)
}

// Get returns the value of the config with an optional default value.
func Get(uri string, d ...string) string {
	v, _ := c.Get(uri, d...)
	return v
}

// TODO: Add basic type getters: Int, Float, etc.

// Set sets the value of the config.
func Set(uri, data string) error {
	return c.Set(uri, data)
}

// All returns all the configs.
func All() []ConfigOption {
	v, err := c.GetAll()
	if err != nil {
		panic(err.Error())
	}

	return v
}

// Unset removes the config.
func Unset(uri string) error {
	return c.Unset(uri)
}
