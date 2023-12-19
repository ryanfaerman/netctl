package config

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	goose "github.com/pressly/goose/v3"

	_ "github.com/glebarez/go-sqlite"

	"github.com/charmbracelet/log"
	dao "github.com/ryanfaerman/netctl/config/data"
)

var (
	c = New("config", ":memory:")

	Logger       = log.Default()
	DatabaseName = "config.db"
)

//go:embed migrations/*.sql
var migrations embed.FS

type config struct {
	name string
	path string

	db      *sql.DB
	queries *dao.Queries
	once    sync.Once
	log     *log.Logger
	pragmas map[string]string

	isLoaded   bool
	loadStatus chan bool
}

func New(name, path string) *config {
	c := &config{
		name: name,
		path: path,
		log:  Logger.With("service", "config"),
		pragmas: map[string]string{
			"journal_mode": "ON",
		},
		loadStatus: make(chan bool),
	}
	c.log.SetLevel(log.DebugLevel)

	return c
}

// WithDefaults establishes known defaults for writing the internal database
func (c *config) WithDefaults() *config {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err.Error())
	}

	executablePath, err := os.Executable()
	if err != nil {
		panic(err.Error())
	}

	applicationName := filepath.Base(executablePath)
	path := filepath.Join(configDir, applicationName, DatabaseName)

	c.name = applicationName
	c.path = path

	return c
}

func (c *config) Name() string { return c.name }

func (c *config) WithName(name string) *config {
	c.name = name
	return c
}

func (c *config) WithPath(path string) *config {
	c.path = path
	return c
}

func (c *config) WithPragma(name, value string) *config {
	c.pragmas[name] = value
	return c
}

// Load the config system, storing data at the given path. This can safely be
// called multiple times.
//
// The underlying config data is stored in SQLite. During the Load, all
// migrations are applied and the queries DAO is setup as well. This should
// ensure that the config database has all the right schema. Migrations are
// embedded into the resulting binary and should be available even without the
// source.
func (c *config) Load() error {
	var err error
	c.log.Debug("loading config")

	c.once.Do(func() {
		err = c.load()
	})

	err = Hook.Dispatch(context.Background(), Definition{c: c})
	if err != nil {
		panic(err.Error())
	}

	c.ScanEnv()

	return err
}

// WaitForLoad blocks until the config system is loaded.
func (c *config) WaitForLoad() {
	if c.isLoaded {
		return
	}

	ch := make(chan bool)

	go func() {
		for {
			select {
			case <-c.loadStatus:
				ch <- true
				return

			}
		}
	}()

	<-ch
}

func (c *config) load() error {
	var err error

	c.log.Debug("loading database", "path", c.path)

	if c.path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(c.path), 0750); err != nil {
			return err
		}
	}

	var dsn strings.Builder

	dsn.WriteString(c.path)

	if len(c.pragmas) > 0 {
		dsn.WriteString("?")
	}
	for pragma, value := range c.pragmas {
		fmt.Fprintf(&dsn, "_pragma=%s(%s)&", pragma, value)
	}

	c.db, err = sql.Open("sqlite", dsn.String())
	if err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	// Setup our migrations here
	goose.SetLogger(logAdapter{*c.log})
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("cannot set dialect: %w", err)
	}

	if err := goose.Up(c.db, "migrations"); err != nil {
		return fmt.Errorf("cannot execute migrations: %w", err)
	}

	c.queries = dao.New(c.db)

	if err == nil {
		c.isLoaded = true

		select {
		case c.loadStatus <- true:
		default:
		}
	}

	return nil
}

func (c *config) ScanEnv() {
	c.log.Debug("scanning environment")

	err := Hook.Dispatch(context.Background(), Definition{c: c})
	if err != nil {
		panic(err.Error())
	}

	var (
		configsField strings.Builder
		flagsField   strings.Builder

		fieldFormat = "%s => %s\n"
	)

	for _, e := range os.Environ() {
		if strings.HasPrefix(strings.ToLower(e), c.name) {
			fields := strings.Split(e, "=")
			if len(fields) != 2 {
				continue
			}

			uri := strings.ToLower(fields[0])
			uri = strings.TrimPrefix(uri, c.name+"_")
			uri = c.unescapeUri(uri)

			rawVal := fields[1]

			c.log.Debug("found env", "uri", uri, "rawVal", rawVal)

			if c.IsFlag(uri) {

				boolVal, err := strconv.ParseBool(rawVal)
				if err != nil {
					c.log.Error("invalid bool", "uri", uri, "rawVal", rawVal, "err", err)
					continue
				}

				if err := c.SetFlag(uri, boolVal); err != nil {
					c.log.Error("cannot set flag", "uri", uri, "err", err)
				}

				fmt.Fprintf(&flagsField, fieldFormat, uri, strconv.FormatBool(boolVal))

				continue
			}

			if c.IsConfig(uri) {
				if err := c.Set(uri, rawVal); err != nil {
					c.log.Error("cannot set config", "uri", uri, "err", err)
					continue
				}

				fmt.Fprintf(&configsField, fieldFormat, uri, rawVal)

				continue
			}

			c.log.Warn("invalid env key", "uri", uri, "key", fields[0])

		}
	}

	if configsField.Len() > 0 {
		c.log.Info("loaded config from environment", "defined", configsField.String())
	}
	if flagsField.Len() > 0 {
		c.log.Info("loaded flags from environment", "defined", flagsField.String())
	}

}

// Reset the configuration DB entirely. This is currently done by just deleting
// the underlying SQLite database.
func (c *config) Reset() error {
	c.log.Debug("resetting the config db")

	c.once = sync.Once{}
	c.isLoaded = false
	c.loadStatus = make(chan bool)

	return os.Remove(c.path)
}

// Set the given config URI to the given value. This will replace the value if
// it is already defined. The database is assumed to be available by others
// calling Load.
func (c *config) Set(uri string, data string) error {
	c.WaitForLoad()

	uri = c.escapeUri(uri)
	return c.queries.SetConfig(context.Background(), dao.SetConfigParams{
		Uri:  uri,
		Data: sql.NullString{String: data, Valid: data != ""},
	})
}

// Define a config URI according to the following rules:
func (c *config) Define(uri string, d ...string) error {
	c.WaitForLoad()

	uri = c.escapeUri(uri)

	fallback := ""
	if len(d) > 0 {
		fallback = d[0]
	}

	return c.queries.DefineConfig(context.Background(), dao.DefineConfigParams{
		Uri: uri,
		Data: sql.NullString{
			String: fallback,
			Valid:  len(d) > 0,
		},
	})
}

func (c *config) IsConfig(uri string) bool {
	c.WaitForLoad()

	uri = c.escapeUri(uri)
	_, err := c.queries.GetConfig(context.Background(), uri)
	return err == nil
}

// Get the data for the given uri. If undefined, an empty string is returned.
// An error should only be returned if there is some underlying system problem.
// The database is assumed to be available by others calling Load.
func (c *config) Get(uri string, d ...string) (string, error) {
	c.WaitForLoad()

	fallback := ""
	if len(d) > 0 {
		fallback = d[0]
	}

	uri = c.escapeUri(uri)

	v, err := c.queries.GetConfig(context.Background(), uri)
	if err != nil {
		if err == sql.ErrNoRows {
			return fallback, nil
		}
		return fallback, err
	}
	if v.Valid {
		return v.String, nil
	}

	return "", nil
}

// Close the config and allow any cleanup to occur as part of shutdown. Once
// closed, the config cannot be used.
func (c *config) Close() error {
	return c.db.Close()
}
