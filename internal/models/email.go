package models

import (
	"time"
)

type Email struct {
	Address string

	IsPrimary    bool
	IsPublic     bool
	IsNotifiable bool
	IsVerified   bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	VerifiedAt time.Time
	ID         int64
}
