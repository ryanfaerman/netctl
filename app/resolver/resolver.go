package resolver

import (
	"database/sql"
	"embed"
	"errors"
	"sync"

	"github.com/charmbracelet/log"
	dao "github.com/ryanfaerman/netctl/app/data"

	"github.com/ryanfaerman/netctl/config"

	"github.com/pressly/goose/v3"

	_ "github.com/glebarez/go-sqlite"
)

type Resolver struct {
	db      *sql.DB
	queries *dao.Queries
	once    sync.Once
	log     *log.Logger
}

func New(logger *log.Logger, migrations embed.FS) (*Resolver, error) {
	var err error
	r := &Resolver{
		log: logger.With("service", "graph"),
	}

	path := config.Get("graph.storage.path")
	if path == "" {
		return r, errors.New("invalid graph.storage.path (empty)")
	}

	r.once.Do(func() {
		r.log.Debug("loading board data", "path", path)

		r.db, err = sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
		if err != nil {
			return
		}

		// Setup our migrations here
		goose.SetLogger(logAdapter{*r.log})

		goose.SetBaseFS(migrations)

		err = goose.SetDialect("sqlite")
		if err != nil {
			return
		}

		err = goose.Up(r.db, "sql/migrations")
		if err != nil {
			return
		}

		r.queries = dao.New(r.db)

	})

	return r, err
}
