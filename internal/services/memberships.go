package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/hamdb"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models"
)

type membership struct{}

var Membership membership

var (
	ErrClubRequiresCallsign   = errors.New("clubs require a callsign")
	ErrCallsignCreationFailed = errors.New("unable to create callsign")
)

func (s membership) Create(ctx context.Context, owner, m *models.Account, callsigns ...string) error {
	if owner.IsAnonymous() {
		return fmt.Errorf("anonymous users cannot create organizations")
	}
	if err := Validation.Apply(m); err != nil {
		return err
	}

	tx, err := global.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := global.dao.WithTx(tx)

	var (
		callsign   string
		callsignID int64
	)
	if m.Kind == models.AccountKindClub {
		if len(callsigns) == 0 {
			return ErrClubRequiresCallsign
		}
		if len(callsigns) > 0 {
			callsign = callsigns[0]
			if callsign == "" {
				return ErrClubRequiresCallsign
			}
		}
	}

	if callsign != "" {
		global.log.Debug("checking if callsign is already associated", "callsign", callsign)
		_, err := qtx.FindAccountByCallsign(ctx, callsign)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		} else {
			return ErrAccountSetupCallsignTaken
		}

		global.log.Debug("validating callsign with hamdb", "callsign", callsign)
		license, err := hamdb.Lookup(ctx, callsign)
		if err != nil {
			return ErrAccountSetupInvalidCallsign
		}
		if license.Class != hamdb.ClubClass {
			return ErrAccountSetupCallsignIndividual
		}

		global.log.Debug("creating callsign record", "callsign", callsign)
		id, err := qtx.CreateCallsignAndReturnId(ctx, dao.CreateCallsignAndReturnIdParams{
			Callsign: license.Call,
			Class:    int64(license.Class),
			Expires: sql.NullTime{
				Time:  license.Expires.Value,
				Valid: license.Expires.Known,
			},
			Status: int64(license.Status),
			Latitude: sql.NullFloat64{
				Float64: license.Lat.Value,
				Valid:   license.Lat.Known,
			},
			Longitude: sql.NullFloat64{
				Float64: license.Lon.Value,
				Valid:   license.Lon.Known,
			},
			Firstname:  sql.NullString{String: license.FirstName, Valid: true},
			Middlename: sql.NullString{String: license.MiddleInitial, Valid: true},
			Lastname:   sql.NullString{String: license.LastName, Valid: true},
			Suffix:     sql.NullString{String: license.Suffix, Valid: true},
			Address:    sql.NullString{String: license.Address, Valid: true},
			City:       sql.NullString{String: license.City, Valid: true},
			State:      sql.NullString{String: license.State, Valid: true},
			Zip:        sql.NullString{String: license.Zip, Valid: true},
			Country:    sql.NullString{String: license.Country, Valid: true},
		})
		if err != nil {
			return ErrCallsignCreationFailed
		}
		callsignID = id
		m.Slug = license.Call
	}

	id, err := qtx.CreateAccount(ctx, dao.CreateAccountParams{
		Name: m.Name,
		Kind: int64(m.Kind),
		Slug: m.Slug,
	})
	if err != nil {
		return fmt.Errorf("error creating account: %w", err)
	}
	m.ID = id
	spew.Dump("new account", m)

	if callsignID != 0 {
		if err := qtx.AssociateCallsignWithAccount(ctx, dao.AssociateCallsignWithAccountParams{
			AccountID:  m.ID,
			CallsignID: callsignID,
		}); err != nil {
			return fmt.Errorf("error associating callsign with account: %w", err)
		}
	}

	roleID, err := qtx.CreateRoleOnAccount(ctx, dao.CreateRoleOnAccountParams{
		Name:        "Owner",
		AccountID:   m.ID,
		Permissions: int64(models.PermissionOwner),
		Ranking:     0,
	})
	if err != nil {
		return fmt.Errorf("error creating role: %w", err)
	}
	fmt.Println("roleID", roleID)

	params := dao.CreateMembershipParams{
		AccountID: owner.ID,
		MemberOf:  m.ID,
		RoleID:    roleID,
	}

	spew.Dump(params)
	err = qtx.CreateMembership(ctx, params)
	if err != nil {
		return fmt.Errorf("error creating membership: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	return nil
}

var ErrAccountSetupCallsignIndividual = fmt.Errorf("callsign is not a club")