package migrations

import (
	"runtime"

	"github.com/pressly/goose/v3"

	"github.com/charmbracelet/log"
)

// A migrations file is a Go file that perform a migration. It must start with
// a number, followed by an underscore and a description of the migration.
type Migration struct {
	Up   goose.GoMigrationContext
	Down goose.GoMigrationContext
	Name string
}

var (
	Migrations []Migration
	Log        *log.Logger
)

// Add a migration to the list of migrations to run
func AddMigration(up, down goose.GoMigrationContext) {
	_, filename, _, _ := runtime.Caller(1)
	Migrations = append(Migrations, Migration{
		Name: filename,
		Up:   up,
		Down: down,
	})
}
