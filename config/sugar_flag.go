package config

import "github.com/ryanfaerman/netctl/hook"

// Flag is a sugar for accessing the flag config options.
var Flag = flagSugar{}

type flagSugar struct{}

// Define registers a new flag, with an optional default value.
func (flagSugar) Define(uri string, d ...bool) {
	Hook.Register(func(e hook.Event[Definition]) {
		e.Payload.Flag(uri, d...)
	})
}

// IsFlag checks if the config uri is a flag.
func (flagSugar) IsFlag(uri string) bool {
	return c.IsFlag(uri)
}

// Get returns the value of the flag.
func (flagSugar) Get(uri string, d ...bool) bool {
	v, err := c.Flag(uri, d...)
	if err != nil {
		panic(err.Error())
	}

	return v
}

// Set sets the value of the flag.
func (flagSugar) Set(uri string, v bool) error {
	return c.SetFlag(uri, v)
}

// All returns all the flags.
func (flagSugar) All() []ConfigOption {
	v, err := c.Flags()
	if err != nil {
		panic(err.Error())
	}
	return v
}
