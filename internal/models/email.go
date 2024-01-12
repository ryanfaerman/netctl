package models

import (
	"time"
)

type Email struct {
	ID int64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	Address      string
	IsPrimary    bool
	IsPublic     bool
	IsNotifiable bool
	VerifiedAt   time.Time
	IsVerified   bool
}
