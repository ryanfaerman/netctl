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

func CurrentAccount(ctx context.Context) *models.Account {
	return services.Session.GetAccount(ctx)
}

func UserCan(ctx context.Context, action string, resources ...any) bool {
	if action == "" {
		return true
	}
	a := services.Session.GetAccount(ctx)
	if err := services.Authorization.Can(ctx, a, action, resources...); err != nil {
		return false
	}
	return true
}
