package models

import (
	"time"
)

type Email struct {
	Address string `form:"email" validate:"required,email"`

	IsVerified bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	VerifiedAt time.Time
	ID         int64
}
