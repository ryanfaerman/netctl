package sql

import (
	"context"
	"database/sql"
	"embed"

	"github.com/charmbracelet/log"
	"github.com/pressly/goose/v3"

	_ "github.com/glebarez/go-sqlite"
	goMigrations "github.com/ryanfaerman/netctl/internal/sql/migrations"
)

//go:generate sqlc generate

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(log *log.Logger, db *sql.DB) error {
	l := log.With("pkg", "sql")
	goMigrations.Log = log.With("pkg", "go-migrations")

	goose.SetLogger(logAdapter{*l})
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	// Add these migrations manually
	for _, migration := range goMigrations.Migrations {
		goose.AddNamedMigrationContext(migration.Name, migration.Up, migration.Down)
	}

	return goose.UpContext(context.Background(), db, "migrations")
}
