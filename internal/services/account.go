package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"dario.cat/mergo"
	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models"
	. "github.com/ryanfaerman/netctl/internal/models/finders"
)

type account struct{}

var Account account

func (account) FindByID(ctx context.Context, id int64) (*models.Account, error) {
	return models.FindAccountByID(ctx, id)
}

func (account) FindByEmail(ctx context.Context, email string) (*models.Account, error) {
	return models.FindAccountByEmail(ctx, email)
}

func (account) FindByCallsign(ctx context.Context, callsign string) (*models.Account, error) {
	return models.FindAccountByCallsign(ctx, callsign)
}

func (account) FindBySlug(ctx context.Context, slug string) (*models.Account, error) {
	return FindOne[models.Account](ctx, BySlug(slug))
}

func (s account) CreateWithEmail(ctx context.Context, email string) (*models.Account, error) {
	if u, err := s.FindByEmail(ctx, email); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	} else {
		return u, nil
	}

	tx, err := global.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := global.dao.WithTx(tx)

	id, err := qtx.CreateAccountAndReturnId(ctx)
	if err != nil {
		return nil, err
	}

	err = qtx.AddPrimaryEmailForAccount(ctx, dao.AddPrimaryEmailForAccountParams{
		AccountID: id,
		Address:   email,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.FindByEmail(ctx, email)
}

var (
	ErrAccountSetupInvalidCallsign = errors.New("invalid callsign")
	ErrAccountSetupCallsignTaken   = errors.New("callsign already taken")
	ErrAccountSetup                = errors.New("unable to setup account")
	ErrAccountSetupCallsignClub    = errors.New("callsign must be for an individual")
)

func (s account) Setup(ctx context.Context, id int64, name, callsign string) error {
	panic("needs re-implementation")
	// account, err := global.dao.UpdateAccount(ctx, dao.UpdateAccountParams{
	// 	ID:   id,
	// 	Name: name,
	// })
	// if err != nil {
	// 	return err
	// }

	// accountCallsign, err := global.dao.FindAccountByCallsign(ctx, callsign)
	// if err != nil {
	// 	if err != sql.ErrNoRows {
	// 		return err
	// 	}
	// } else {
	// 	if account.ID != accountCallsign.ID {
	// 		return ErrAccountSetupCallsignTaken
	// 	}
	// }
	//
	// fccCallsign, err := hamdb.Lookup(ctx, callsign)
	// if err != nil {
	// 	return ErrAccountSetupInvalidCallsign
	// }
	// if fccCallsign.Class == hamdb.ClubClass {
	// 	return ErrAccountSetupCallsignClub
	// }
	//
	// tx, err := global.db.BeginTx(ctx, nil)
	// if err != nil {
	// 	return err
	// }
	// defer tx.Rollback()
	//
	// qtx := global.dao.WithTx(tx)
	//
	// callsignID, err := qtx.CreateCallsignAndReturnId(ctx, dao.CreateCallsignAndReturnIdParams{
	// 	Callsign: fccCallsign.Call,
	// 	Class:    int64(fccCallsign.Class),
	// 	Expires: sql.NullTime{
	// 		Time:  fccCallsign.Expires.Value,
	// 		Valid: fccCallsign.Expires.Known,
	// 	},
	// 	Status: int64(fccCallsign.Status),
	// 	Latitude: sql.NullFloat64{
	// 		Float64: fccCallsign.Lat.Value,
	// 		Valid:   fccCallsign.Lat.Known,
	// 	},
	// 	Longitude: sql.NullFloat64{
	// 		Float64: fccCallsign.Lon.Value,
	// 		Valid:   fccCallsign.Lon.Known,
	// 	},
	// 	Firstname:  sql.NullString{String: fccCallsign.FirstName, Valid: true},
	// 	Middlename: sql.NullString{String: fccCallsign.MiddleInitial, Valid: true},
	// 	Lastname:   sql.NullString{String: fccCallsign.LastName, Valid: true},
	// 	Suffix:     sql.NullString{String: fccCallsign.Suffix, Valid: true},
	// 	Address:    sql.NullString{String: fccCallsign.Address, Valid: true},
	// 	City:       sql.NullString{String: fccCallsign.City, Valid: true},
	// 	State:      sql.NullString{String: fccCallsign.State, Valid: true},
	// 	Zip:        sql.NullString{String: fccCallsign.Zip, Valid: true},
	// 	Country:    sql.NullString{String: fccCallsign.Country, Valid: true},
	// })
	// if err != nil {
	// 	return errors.New("unable to create callsign")
	// }
	//
	// if err := qtx.AssociateCallsignWithAccount(ctx, dao.AssociateCallsignWithAccountParams{
	// 	AccountID:  account.ID,
	// 	CallsignID: callsignID,
	// }); err != nil {
	// 	return err
	// }
	//
	// return tx.Commit()
}

func (s account) AvatarURL(ctx context.Context, callsigns ...string) string {
	var err error
	account := Session.GetAccount(ctx)
	if len(callsigns) > 0 {
		if account.Callsign().Call != callsigns[0] {
			account, err = s.FindByCallsign(ctx, callsigns[0])
			if err != nil {
				return ""
			}
		}
	}
	email, err := account.PrimaryEmail()
	if err != nil {
		return ""
	}

	h := sha256.New()
	h.Write([]byte(strings.TrimSpace(strings.ToLower(email.Address))))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x", h.Sum(nil))
}

func (s account) Setting(ctx context.Context, path string) any {
	account := Session.GetAccount(ctx)
	val := account.Setting(ctx, path)
	global.log.Warn("using setting", "path", path, "val", val)

	return val
}

func (a account) SaveSettings(ctx context.Context, id int64, settings *models.Settings) error {
	account, err := a.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := mergo.Merge(&account.Settings, settings, mergo.WithOverride); err != nil {
		return err
	}

	fmt.Println("right before validation")
	spew.Dump(account.Settings)
	if err := Validation.Apply(account.Settings); err != nil {
		return err
	}

	data, err := json.Marshal(account.Settings)
	if err != nil {
		return err
	}
	fmt.Println("right before save")
	spew.Dump(data)
	spew.Dump(account.Settings)

	if err := global.dao.UpdateAccountSettings(ctx, dao.UpdateAccountSettingsParams{
		ID:       account.ID,
		Settings: string(data),
	}); err != nil {
		return err
	}
	fmt.Println("right after save")
	spew.Dump(account)

	Session.ClearAccountCache(ctx, account)

	return nil
}

func (s account) Geolocation(ctx context.Context, m *models.Account) (float64, float64, error) {
	call := m.Callsign()
	return call.Latitude, call.Longitude, nil
}
