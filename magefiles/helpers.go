package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/log"
	"github.com/magefile/mage/sh"
)

func NewMigration(name string) error {
	migrations, err := ioutil.ReadDir("internal/sql/migrations")
	if err != nil {
		return err
	}

	index := 0
	for _, migration := range migrations {
		if migration.Name() == ".DS_Store" {
			continue
		}
		parts := strings.SplitN(migration.Name(), "_", 2)
		migrationIndex, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		if migrationIndex > index {
			index = migrationIndex
		}
	}

	{
		var result []rune

		for _, char := range name {
			if unicode.IsLetter(char) {
				result = append(result, char)
			} else {
				result = append(result, '_')
			}
		}
		name = string(result)
	}

	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("-- %s\n", name))
	b.WriteString("-- +goose Up\n")
	b.WriteString("-- +goose StatementBegin\n\n")
	b.WriteString("-- +goose StatementEnd\n\n")
	b.WriteString("-- +goose Down\n")
	b.WriteString("-- +goose StatementBegin\n\n")
	b.WriteString("-- +goose StatementEnd\n\n")

	filename := fmt.Sprintf("%04d_%s.sql", index+1, name)
	log.Info("Creating migration", "filename", filename)

	return os.WriteFile(filepath.Join("internal/sql/migrations", filename), b.Bytes(), 0644)
}

func CheckUpdates() error {
	log.Info("Checking for updates... (this might take a while)", "start", "starting")
	started := time.Now()
	// go list -m -u -f '{{if not (or .Indirect .Main)}}{{.Update}}{{end}}' all
	output, err := sh.Output("go", "list", "-u", "-m", "-f", "'{{if not (or .Indirect .Main)}}{{.Version}} {{.Update}}{{end}}'", "all")
	var b strings.Builder
	for _, l := range strings.Split(output, "\n") {
		if l == "''" {
			continue
		}
		if l == "'<nil>'" {
			continue
		}
		l = strings.Trim(l, "'")
		parts := strings.SplitN(l, " ", 3)
		if len(parts) != 3 {
			continue
		}
		b.WriteString(fmt.Sprintf("%s: %s -> %s", parts[1], parts[0], parts[2]))
		b.WriteString("\n")
	}

	log.Info("Update check complete", "state", "complete", "elapsed", time.Since(started).String(), "results", b.String())
	return err
}
