package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/mrz1836/postmark"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/internal/models"
)

type session struct{}

var Session session

type sessionPayload struct {
	Token string
	Email string
}

func init() {
	gob.Register(sessionPayload{})
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
			// http.Redirect(w, r, named.RouteURL("user-login"), http.StatusFound)
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

	u, err := User.CreateWithEmail(ctx, payload.Email)
	if err != nil {
		return err
	}

	global.session.Put(ctx, "user_id", u.ID)

	return nil
}

func (session) Destroy(ctx context.Context) error {
	global.log.Info("session destroyed")
	global.session.Clear(ctx)
	return nil
}

var (
	ErrNoUserInSession = errors.New("no user in session")
)

func (session) GetUser(ctx context.Context) (*models.User, error) {
	id, ok := global.session.Get(ctx, "user_id").(int64)
	if !ok {
		return nil, ErrNoUserInSession
	}
	return User.FindByID(ctx, id)
}

func (s session) MustGetUser(ctx context.Context) *models.User {
	u, err := s.GetUser(ctx)
	if err != nil {
		panic(err)
	}
	spew.Dump(u)
	spew.Dump(u.Ready())
	return u
}
