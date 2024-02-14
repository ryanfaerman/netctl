package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/events"
	"github.com/ryanfaerman/netctl/internal/models"

	"github.com/mrz1836/postmark"
	"github.com/ryanfaerman/netctl/config"

	. "github.com/ryanfaerman/netctl/internal/models/finders"
)

type email struct{}

var Email email

func init() {
	Event.Register(events.AccountEmailAdded{}, Email.SendEmailAdditionVerification)
	Event.Register(events.AccountEmailVerified{}, Email.FinalizeEmailChange)
	gob.Register(verificationPayload{})
}

type verificationPayload struct {
	AccountID int64
	EmailID   int64
	Email     string
}

func (email) SendEmailAdditionVerification(ctx context.Context, event models.Event) error {
	if e, ok := event.Event.(*events.AccountEmailAdded); ok {
		spew.Dump(event)
		spew.Dump(e)

		payload := verificationPayload{
			AccountID: event.AccountID,
			EmailID:   e.ID,
			Email:     e.Email,
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

		user, err := FindOne[models.Account](ctx, ByID(event.AccountID))
		if err != nil {
			return err
		}

		global.log.Info("Sending email change verification", "email", e.Email, "token", verifyToken)

		serverToken := config.Get("service.email.server.token")
		accountToken := config.Get("service.email.account.token")
		client := postmark.NewClient(serverToken, accountToken)
		msg := postmark.TemplatedEmail{
			TemplateAlias: "confirm-email-change",
			TemplateModel: map[string]interface{}{
				"product_url": "http://localhost:8090",

				"new_email":    e.Email,
				"product_name": config.Get("service.email.product.name"),
				"action_url":   fmt.Sprintf("http://localhost:8090/settings/%s/emails/-/verify?token=%s", user.Slug, verifyToken),
			},
			From: "bartender@toot.beer",
			To:   e.Email,
			Tag:  "email-addition-verification",
		}
		_, err = client.SendTemplatedEmail(ctx, msg)
		return err
	}
	fmt.Printf("invalid event: %T\n", event)
	global.log.Debug("invalid event2", "event", event, "type", fmt.Sprintf("%T", event))
	return fmt.Errorf("invalid event2 %T", event.Event)
}

func (email) FinalizeEmailChange(ctx context.Context, event models.Event) error {
	if e, ok := event.Event.(*events.AccountEmailVerified); ok {

		account, err := FindOne[models.Account](ctx, ByEmail(e.Email))
		if err != nil {
			return fmt.Errorf("unable to find account by email %s: %w", e.Email, err)
		}

		emails, err := Find[models.Email](ctx, ByAccount(account.ID))
		if err != nil {
			global.log.Error("unable to find emails for account", "error", err)
			return nil
		}
		for _, m := range emails {
			if !m.IsVerified {
				if err := global.dao.DeleteEmail(ctx, e.ID); err != nil {
					return err
				}
				continue
			}

			serverToken := config.Get("service.email.server.token")
			accountToken := config.Get("service.email.account.token")
			client := postmark.NewClient(serverToken, accountToken)
			msg := postmark.TemplatedEmail{
				TemplateAlias: "email-did-change",
				TemplateModel: map[string]interface{}{
					"product_url":  "http://localhost:8090",
					"new_email":    e.Email,
					"product_name": config.Get("service.email.product.name"),
				},
				From: "bartender@toot.beer",
				To:   m.Address,
				Tag:  "email-addition-verification",
			}
			_, err = client.SendTemplatedEmail(ctx, msg)

			if m.Address != e.Email {
				if err := global.dao.DeleteEmail(ctx, m.ID); err != nil {
					return err
				}
			}
		}
		return nil

	}

	fmt.Printf("invalid event: %T\n", event)
	global.log.Debug("invalid event3", "event", event, "type", fmt.Sprintf("%T", event.Event))
	return fmt.Errorf("invalid event3 %T", event.Event)
}

func (email) VerifyEmailAddition(ctx context.Context, account *models.Account, token string) error {
	decoded, err := global.brc.DecodeString(token)
	if err != nil {
		global.log.Error("unable to decode token", "error", err)
		return err
	}

	var payload verificationPayload
	if err := gob.NewDecoder(bytes.NewReader(decoded.Payload())).Decode(&payload); err != nil {
		global.log.Error("unable to decode payload", "error", err)
		return err
	}

	if decoded.IsExpired(600) {
		global.log.Warn("received expired verification token", "token", token, "timestamp", decoded.Timestamp())
		return ErrTokenExpired
	}

	if payload.AccountID != account.ID {
		global.log.Warn("received mismatched account id", "token", token, "account_id", account.ID)
		return ErrTokenMismatch
	}

	eml, err := FindOne[models.Email](ctx, ByID(payload.EmailID))
	if err != nil {
		global.log.Warn("received mismatched email id", "id", payload.EmailID, "error", err)
		return ErrTokenMismatch
	}
	if eml.Address != payload.Email {
		global.log.Warn("received mismatched email", "email", payload.Email, "address", eml.Address)
		return ErrTokenMismatch
	}

	if err := global.dao.SetEmailVerified(ctx, payload.EmailID); err != nil {
		global.log.Error("unable to set email verified", "error", err)
		return err
	}

	if err := Event.Create(ctx, account.StreamID, events.AccountEmailVerified{
		ID:    payload.EmailID,
		Email: payload.Email,
	}); err != nil {
		global.log.Error("unable to create email verified event", "error", err)
		return err
	}

	global.log.Info("Verifying email change", "payload", payload)
	spew.Dump(payload)
	return nil
}
