package app

import (
	"os"
	"path/filepath"

	"github.com/ryanfaerman/netctl/config"
)

func init() {
	userCache, err := os.UserCacheDir()
	if err != nil {
		panic(err.Error())
	}

	config.Define("graph.storage.path", filepath.Join(userCache, "retro", "board.db"))
	config.Flag.Define("graph.playground", true)
}
