package services

import (
	"context"
	"database/sql"

	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/models"
)

type user struct{}

var User user

func (user) FindByID(ctx context.Context, id int64) (*models.User, error) {
	raw, err := global.dao.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	u := models.User{
		ID:        raw.ID,
		Name:      raw.Name,
		CreatedAt: raw.Createdat,
	}

	if raw.Deletedat.Valid {
		u.DeletedAt = raw.Deletedat.Time
		u.Deleted = true
	}

	return &u, nil

}

func (user) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	raw, err := global.dao.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	u := models.User{
		ID:        raw.ID,
		Name:      raw.Name,
		CreatedAt: raw.Createdat,
	}

	if raw.Deletedat.Valid {
		u.DeletedAt = raw.Deletedat.Time
		u.Deleted = true
	}
	return &u, nil
}

func (s user) CreateWithEmail(ctx context.Context, email string) (*models.User, error) {
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

	id, err := qtx.CreateUserAndReturnId(ctx)
	if err != nil {
		return nil, err
	}

	err = qtx.AddPrimaryEmailForUser(ctx, dao.AddPrimaryEmailForUserParams{
		UserID:  id,
		Address: email,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.FindByEmail(ctx, email)

}
