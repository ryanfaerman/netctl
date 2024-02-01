package views

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"

	"github.com/essentialkaos/branca/v2"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
)

var (
	brc  branca.Branca
	err  error
	once sync.Once
)

func init() {
	gob.Register(InputAttrs{})
}

func (i InputAttrs) Encode() string {
	once.Do(func() {
		brc, err = branca.NewBranca([]byte(config.Get("random_key")))
		if err != nil {
			panic("unable to create branca")
		}
	})

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(i); err != nil {
		panic("unable to encode payload")
	}

	encoded, err := brc.EncodeToString(b.Bytes())
	if err != nil {
		return ""
	}
	return encoded
}

func DecodeInputAttrs(encoded string) (InputAttrs, error) {
	once.Do(func() {
		brc, err = branca.NewBranca([]byte(config.Get("random_key")))
		if err != nil {
			panic("unable to create branca")
		}
	})

	decoded, err := brc.DecodeString(encoded)
	if err != nil {
		return InputAttrs{}, err
	}

	var i InputAttrs
	if err := gob.NewDecoder(bytes.NewReader(decoded.Payload())).Decode(&i); err != nil {
		return InputAttrs{}, err
	}
	return i, nil
}

func selfGravatarURL(ctx context.Context) string {
	u, err := services.Account.AvatarURLForAccount(ctx)
	if err != nil {
		return ""
	}
	return u
}

func callsignGravatarURL(ctx context.Context, callsign string) string {
	u, err := services.Account.AvatarURLForCallsign(ctx, callsign)
	if err != nil {
		return ""
	}
	return u
}

func CurrentAccount(ctx context.Context) *models.Account {
	m, _ := services.Session.GetAccount(ctx)
	return m
}
