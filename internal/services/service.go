package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

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
	ctxKeyTX
)

var setupOnce sync.Once

func SetupSSE(s *sse.Server) {
	global.events = s
}

func Setup(logger *log.Logger, db *sql.DB) error {
	var err error

	setupOnce.Do(func() {
		global.log = logger.With("pkg", "services")
		global.log.Debug("running setup tasks")

		global.db = db
		global.dao = dao.New(global.db)

		global.brc, err = branca.NewBranca([]byte(config.Get("random_key")))
		if err != nil {
			return
		}

		err = models.Setup(logger, db)

		lifetimeStr := config.Get("session.lifetime", "240h")
		lifetime, err := time.ParseDuration(lifetimeStr)
		if err != nil {
			panic(fmt.Sprintf("unable to parse session.lifetime: %s", err))
		}

		global.session.Store = sqlite3store.New(global.db)
		global.session.Lifetime = lifetime
		global.session.Cookie.Name = config.Get("session.name", "_session")
		global.session.Cookie.Path = config.Get("session.path", "/")
	})

	return err
}

func database(ctx context.Context) *sql.DB {
	return nil
}

func transaction(ctx context.Context, fn func(context.Context, *dao.Queries) error) error {
	var (
		tx  *sql.Tx
		err error
		ok  bool
	)
	tx, ok = ctx.Value(ctxKeyTX).(*sql.Tx)
	if !ok {
		fmt.Println("starting transaction")
		tx, err = global.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, ctxKeyTX, tx)
		defer tx.Rollback()
	}
	fmt.Println("are we ok?", ok)

	err = fn(ctx, global.dao.WithTx(tx))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("committing")
		return tx.Commit()
	}
	return nil
}
