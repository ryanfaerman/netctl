// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0

package dao

import (
	"database/sql"
	"time"
)

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
	UserID       int64
	Address      string
	Isprimary    bool
	Ispublic     bool
	Isnotifiable bool
	Verifiedat   sql.NullTime
}

type Session struct {
	Token  string
	Data   []byte
	Expiry float64
}

type User struct {
	ID        int64
	Name      string
	Createdat time.Time
	Updatedat time.Time
	Deletedat sql.NullTime
}

type UsersCallsign struct {
	UserID     int64
	CallsignID int64
}

type UsersSession struct {
	UserID    int64
	Token     string
	Createdat time.Time
	Createdby string
}
