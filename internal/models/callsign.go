package models

import (
	"time"
)

type Callsign struct {
	ID int64

	Call       string
	Class      int64
	Expires    time.Time
	Status     int64
	Latitude   float64
	Longitude  float64
	Firstname  string
	Middlename string
	Lastname   string
	Suffix     string
	Address    string
	City       string
	State      string
	Zip        string
	Country    string

	CreatedAt time.Time
	UpdatedAt time.Time
}
