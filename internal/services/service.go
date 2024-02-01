package services

import (
	"database/sql"
	"sync"

	scs "github.com/alexedwards/scs/v2"
	"github.com/r3labs/sse/v2"

	"github.com/alexedwards/scs/sqlite3store"

	branca "github.com/essentialkaos/branca/v2"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models"

	"github.com/charmbracelet/log"
)

var global = struct {
	session *scs.SessionManager
	db      *sql.DB
	dao     *dao.Queries
	log     *log.Logger
	brc     branca.Branca
	events  *sse.Server
}{
	session: scs.New(),
	log:     log.With("pkg", "services"),
}

func init() {
	config.Define("service.email.account.token")
	config.Define("service.email.server.token")
	config.Define("service.email.product.name", "Net Control")
	config.Define("service.email.product.url", "http://localhost:8090")
}

type ctxKey int

const (
	ctxKeyCSRF ctxKey = iota
	ctxKeyUser
	ctxKeyAccount
)

var setupOnce sync.Once

func SetupSSE(s *sse.Server) {
	global.events = s
}

func Setup(logger *log.Logger, db *sql.DB) error {
	var err error

	setupOnce.Do(func() {
		global.log = logger
		global.log = log.With("pkg", "services")
		global.log.Debug("running setup tasks")

		global.db = db
		global.dao = dao.New(global.db)

		global.brc, err = branca.NewBranca([]byte(config.Get("random_key")))
		if err != nil {
			return
		}

		err = models.Setup(logger, db)

		global.session.Store = sqlite3store.New(global.db)

		global.session.Cookie.Name = config.Get("session.name", "_session")
		global.session.Cookie.Path = config.Get("session.path", "/")
	})

	return err
}
