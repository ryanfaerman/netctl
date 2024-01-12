package sql

import (
	"database/sql"
	"embed"

	"github.com/charmbracelet/log"
	"github.com/pressly/goose/v3"
)

//go:generate sqlc generate

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(log *log.Logger, db *sql.DB) error {
	l := log.With("pgk", "sql")

	goose.SetLogger(logAdapter{*l})
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}
