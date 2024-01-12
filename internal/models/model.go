package models

import (
	"database/sql"
	"sync"

	"github.com/ryanfaerman/netctl/internal/dao"

	"github.com/charmbracelet/log"
	_ "github.com/glebarez/go-sqlite"

	migrations "github.com/ryanfaerman/netctl/internal/sql"
)

var global = struct {
	db  *sql.DB
	log *log.Logger
	dao *dao.Queries
}{
	log: log.With("pkg", "models"),
}

func init() {
}

var setupOnce sync.Once

func Setup(logger *log.Logger, db *sql.DB) error {
	var err error

	setupOnce.Do(func() {
		global.log = logger.With("pkg", "models")
		global.log.Debug("running setup tasks")
		global.db = db

		err = migrations.RunMigrations(global.log, global.db)
		if err != nil {
			return
		}

		global.dao = dao.New(global.db)

	})

	return err
}
