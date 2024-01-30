package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/log"
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
			return err
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
