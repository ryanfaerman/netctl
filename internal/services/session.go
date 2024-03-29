package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"github.com/mrz1836/postmark"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/internal/models"

	ttlcache "github.com/jellydator/ttlcache/v3"
)

type session struct {
	accountCache *ttlcache.Cache[int64, *models.Account]
}

var Session = session{
	accountCache: ttlcache.New[int64, *models.Account](
		ttlcache.WithTTL[int64, *models.Account](global.session.Lifetime),
	),
}

type sessionPayload struct {
	Token string
	Email string
}

func init() {
	gob.Register(sessionPayload{})
	go Session.accountCache.Start()
}

func (session) IsAuthenticated(ctx context.Context) bool {
	return global.session.GetBool(ctx, "authenticated")
}

func (session) Middleware(next http.Handler) http.Handler {
	return global.session.LoadAndSave(next)
}

func (session) AuthenticatedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !global.session.GetBool(r.Context(), "authenticated") {
			// http.Redirect(w, r, named.RouteURL("Account-login"), http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (session) SendEmailVerification(ctx context.Context, email string) error {
	payload := sessionPayload{
		Token: global.session.Token(ctx),
		Email: email,
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(payload); err != nil {
		global.log.Error("unable to encode payload", "error", err)
		return err
	}

	verifyToken, err := global.brc.EncodeToString(b.Bytes())
	if err != nil {
		global.log.Error("unable to generate verification token", "error", err)
		return err
	}

	global.log.Info("Sending email verification", "email", email, "verify_token", verifyToken)
	serverToken := config.Get("service.email.server.token")
	accountToken := config.Get("service.email.account.token")
	client := postmark.NewClient(serverToken, accountToken)
	msg := postmark.TemplatedEmail{
		TemplateAlias: "magic-link",
		TemplateModel: map[string]interface{}{
			"product_url":  "http://localhost:8090",
			"product_name": config.Get("service.email.product.name"),
			"action_url":   fmt.Sprintf("http://localhost:8090/session/verify?token=%s", verifyToken),
		},
		From: "bartender@toot.beer",
		To:   email,
		Tag:  "magic-sign-in",
	}
	_, err = client.SendTemplatedEmail(ctx, msg)
	return err
}

var (
	ErrTokenInvalid  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenMismatch = errors.New("token mismatch")
	ErrTokenDecode   = errors.New("token decode failed")
)

func (session) Verify(ctx context.Context, token string) error {
	if len(token) == 0 {
		return ErrTokenInvalid
	}

	decoded, err := global.brc.DecodeString(token)
	if err != nil {
		global.log.Error("unable to decode verification token", "token", token, "error", err)
		return err
	}

	fmt.Println(decoded.Timestamp())
	if decoded.IsExpired(600) {
		global.log.Warn("received expired verification token", "token", token, "timestamp", decoded.Timestamp())
		return ErrTokenExpired
	}

	var payload sessionPayload
	if err := gob.NewDecoder(bytes.NewReader(decoded.Payload())).Decode(&payload); err != nil {
		global.log.Error("unable to decode verification payload", "token", token, "error", err)
		return ErrTokenDecode
	}

	if payload.Token != global.session.Token(ctx) {
		global.log.Warn("received mismatched session token", "token", token, "session_token", global.session.Token(ctx))
		return ErrTokenMismatch
	}

	global.session.Put(ctx, "authenticated", true)
	global.log.Info("session verified", "email", payload.Email)

	u, err := Account.CreateWithEmail(ctx, payload.Email)
	if err != nil {
		return err
	}

	global.session.Put(ctx, "account", u)
	global.session.Put(ctx, "account_id", u.ID)

	return nil
}

func (session) Destroy(ctx context.Context) error {
	global.log.Info("session destroyed")
	global.session.Clear(ctx)
	return nil
}

var ErrNoAccountInSession = errors.New("no account in session")

func (s session) GetAccount(ctx context.Context) *models.Account {
	id, ok := global.session.Get(ctx, "account_id").(int64)
	if !ok {
		return models.AccountAnonymous
	}
	if id < 0 {
		return models.AccountAnonymous
	}

	if s.accountCache.Has(id) {
		global.log.Debug("getting account from cache")
		account := s.accountCache.Get(id).Value()
		account.Cached = true
		return account
	}

	account, err := models.FindAccountByID(ctx, id)
	if err != nil {
		if err != sql.ErrNoRows {
			global.log.Error("unable to find account by id", "id", id, "error", err)
		}
		return models.AccountAnonymous
	}
	s.accountCache.Set(id, account, ttlcache.DefaultTTL)
	return account
}

func (s session) SetAccount(ctx context.Context, account *models.Account) {
	global.session.Put(ctx, "account_id", account.ID)
	s.accountCache.Delete(account.ID)
}

func (s session) ClearAccountCache(ctx context.Context, account *models.Account) {
	s.accountCache.Delete(account.ID)
}
