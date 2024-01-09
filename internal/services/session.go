package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
)

type Session struct {
	Token string
	Email string
}

func init() {
	gob.Register(Session{})
}

func (Session) IsAuthenticated(ctx context.Context) bool {
	return global.session.GetBool(ctx, "authenticated")
}

func (Session) Middleware(next http.Handler) http.Handler {
	return global.session.LoadAndSave(next)
}

func (Session) SendEmailVerification(ctx context.Context, email string) error {
	payload := Session{
		Token: global.session.Token(ctx),
		Email: email,
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(payload); err != nil {
		global.log.Error("unable to encode payload", "error", err)
		return err
	}

	verifyToken, err := brc().EncodeToString(b.Bytes())
	if err != nil {
		global.log.Error("unable to generate verification token", "error", err)
		return err
	}

	global.log.Info("Sending email verification", "email", email, "verify_token", verifyToken)
	return nil
}

func (Session) Verify(ctx context.Context, token string) error {
	if len(token) == 0 {
		return errors.New("token verification failed; invalid token")
	}

	decoded, err := global.brc.DecodeString(token)
	if err != nil {
		return err
	}

	fmt.Println(decoded.Timestamp())
	if decoded.IsExpired(60) {
		global.log.Warn("received expired verification token", "token", token, "timestamp", decoded.Timestamp())
		return errors.New("token verification failed; token expired")
	}

	var payload Session
	if err := gob.NewDecoder(bytes.NewReader(decoded.Payload())).Decode(&payload); err != nil {
		global.log.Error("unable to decode verification payload", "token", token, "error", err)
		return errors.New("token verification failed; unable to decode")
	}

	if payload.Token != global.session.Token(ctx) {
		global.log.Warn("received mismatched session token", "token", token, "session_token", global.session.Token(ctx))
		return errors.New("token verification failed; mismatched session token")
	}

	global.session.Put(ctx, "authenticated", true)
	global.log.Info("session verified", "email", payload.Email)

	return nil
}

func (Session) Destroy(ctx context.Context) error {
	global.log.Info("session destroyed")
	global.session.Clear(ctx)
	return nil
}
