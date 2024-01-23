// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package dao

import (
	"database/sql"
	"time"
)

type Account struct {
	ID        int64
	Name      string
	Createdat time.Time
	Updatedat time.Time
	Deletedat sql.NullTime
	Kind      int64
}

type AccountsCallsign struct {
	AccountID  int64
	CallsignID int64
}

type AccountsSession struct {
	AccountID int64
	Token     string
	Createdat time.Time
	Createdby string
}

type Callsign struct {
	ID         int64
	Createdat  time.Time
	Updatedat  time.Time
	Callsign   string
	Class      int64
	Expires    sql.NullTime
	Status     int64
	Grid       sql.NullString
	Latitude   sql.NullFloat64
	Longitude  sql.NullFloat64
	Firstname  sql.NullString
	Middlename sql.NullString
	Lastname   sql.NullString
	Suffix     sql.NullString
	Address    sql.NullString
	City       sql.NullString
	State      sql.NullString
	Zip        sql.NullString
	Country    sql.NullString
}

type Email struct {
	ID           int64
	Createdat    time.Time
	Updatedat    time.Time
	AccountID    int64
	Address      string
	Isprimary    bool
	Ispublic     bool
	Isnotifiable bool
	Verifiedat   sql.NullTime
}

type Event struct {
	ID        int64
	Created   time.Time
	StreamID  string
	AccountID int64
	EventType string
	EventData []byte
}

type Net struct {
	ID      int64
	Name    string
	Created time.Time
	Updated time.Time
	Deleted sql.NullTime
}

type NetSession struct {
	ID       int64
	NetID    int64
	StreamID string
	Created  time.Time
}

type Session struct {
	Token  string
	Data   []byte
	Expiry float64
}