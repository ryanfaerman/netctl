package services

import (
	"database/sql"

	scs "github.com/alexedwards/scs/v2"
	branca "github.com/essentialkaos/branca/v2"
	"github.com/ryanfaerman/netctl/config"

	"github.com/charmbracelet/log"
)

var global = struct {
	session *scs.SessionManager
	db      *sql.DB
	log     *log.Logger
	brc     branca.Branca
}{
	session: scs.New(),
	log:     log.With("service", "session"),
}

func brc() branca.Branca {
	if global.brc == nil {
		if br, err := branca.NewBranca([]byte(config.Get("random_key"))); err != nil {
			panic(err)
		} else {
			global.brc = br

		}
	}
	return global.brc
}

type ctxKey int

const (
	ctxKeyCSRF ctxKey = iota
)

func SetDatabase(db *sql.DB) { global.db = db }

func RunMigrations() error {
	return nil
}

func SetLogger(log *log.Logger) { global.log = log }

func ConfigureSessionManager(fn func(*scs.SessionManager)) { fn(global.session) }
