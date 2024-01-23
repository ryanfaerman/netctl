package models

import (
	"time"
)

type Callsign struct {
	ID        int64
	Call      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
