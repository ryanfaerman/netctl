// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: callsigns.sql

package dao

import (
	"context"
	"database/sql"
)

const associateCallsignWithAccount = `-- name: AssociateCallsignWithAccount :exec
INSERT INTO accounts_callsigns (
  account_id, callsign_id
) VALUES (
  ?1, ?2
)
`

type AssociateCallsignWithAccountParams struct {
	AccountID  int64
	CallsignID int64
}

func (q *Queries) AssociateCallsignWithAccount(ctx context.Context, arg AssociateCallsignWithAccountParams) error {
	_, err := q.db.ExecContext(ctx, associateCallsignWithAccount, arg.AccountID, arg.CallsignID)
	return err
}

const createCallsignAndReturnId = `-- name: CreateCallsignAndReturnId :one
INSERT INTO callsigns (
  createdAt, updatedAt,
  callsign,
  class,
  expires,
  status,
  grid,
  latitude,
  longitude,
  firstName,
  middleName,
  lastName,
  suffix,
  address,
  city,
  state,
  zip,
  country
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
  ?1,
  ?2,
  ?3,
  ?4,
  ?5,
  ?6,
  ?7,
  ?8,
  ?9,
  ?10,
  ?11,
  ?12,
  ?13,
  ?14,
  ?15,
  ?16
)
RETURNING id
`

type CreateCallsignAndReturnIdParams struct {
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

func (q *Queries) CreateCallsignAndReturnId(ctx context.Context, arg CreateCallsignAndReturnIdParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createCallsignAndReturnId,
		arg.Callsign,
		arg.Class,
		arg.Expires,
		arg.Status,
		arg.Grid,
		arg.Latitude,
		arg.Longitude,
		arg.Firstname,
		arg.Middlename,
		arg.Lastname,
		arg.Suffix,
		arg.Address,
		arg.City,
		arg.State,
		arg.Zip,
		arg.Country,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const findCallsign = `-- name: FindCallsign :one
SELECT id, createdat, updatedat, callsign, class, expires, status, grid, latitude, longitude, firstname, middlename, lastname, suffix, address, city, state, zip, country
FROM callsigns
WHERE callsigns.callsign = ?1
`

func (q *Queries) FindCallsign(ctx context.Context, callsign string) (Callsign, error) {
	row := q.db.QueryRowContext(ctx, findCallsign, callsign)
	var i Callsign
	err := row.Scan(
		&i.ID,
		&i.Createdat,
		&i.Updatedat,
		&i.Callsign,
		&i.Class,
		&i.Expires,
		&i.Status,
		&i.Grid,
		&i.Latitude,
		&i.Longitude,
		&i.Firstname,
		&i.Middlename,
		&i.Lastname,
		&i.Suffix,
		&i.Address,
		&i.City,
		&i.State,
		&i.Zip,
		&i.Country,
	)
	return i, err
}

const findCallsignsForAccount = `-- name: FindCallsignsForAccount :many
SELECT callsigns.id, callsigns.createdat, callsigns.updatedat, callsigns.callsign, callsigns.class, callsigns.expires, callsigns.status, callsigns.grid, callsigns.latitude, callsigns.longitude, callsigns.firstname, callsigns.middlename, callsigns.lastname, callsigns.suffix, callsigns.address, callsigns.city, callsigns.state, callsigns.zip, callsigns.country
FROM callsigns
JOIN accounts_callsigns ON callsigns.id = accounts_callsigns.callsign_id
WHERE accounts_callsigns.account_id = ?1
`

func (q *Queries) FindCallsignsForAccount(ctx context.Context, accountID int64) ([]Callsign, error) {
	rows, err := q.db.QueryContext(ctx, findCallsignsForAccount, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Callsign
	for rows.Next() {
		var i Callsign
		if err := rows.Scan(
			&i.ID,
			&i.Createdat,
			&i.Updatedat,
			&i.Callsign,
			&i.Class,
			&i.Expires,
			&i.Status,
			&i.Grid,
			&i.Latitude,
			&i.Longitude,
			&i.Firstname,
			&i.Middlename,
			&i.Lastname,
			&i.Suffix,
			&i.Address,
			&i.City,
			&i.State,
			&i.Zip,
			&i.Country,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
