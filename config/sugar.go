package config

import (
	"os"
	"path/filepath"
)

// Name returns the name of the executable.
func Name() string {
	s, _ := os.Executable()
	return filepath.Base(s)
}

// Load loads the default config, stored in the default directory
func Load() error {
	c = c.WithDefaults()

	return c.Load()
}

// LoadFrom loads the config from the specified path. Path must be the absolyte
// path to the config database.
func LoadFrom(path string) error {
	c = c.WithDefaults().WithPath(path)

	return c.Load()
}

// LoadEnv loads the config from the environment variables.
func LoadEnv() {
	c.ScanEnv()
}

// Reset resets the config to the default values.
func Reset() error { return c.Reset() }

// Close disconnects from the config database
func Close() error { return c.Close() }
