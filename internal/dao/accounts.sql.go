// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: accounts.sql

package dao

import (
	"context"
)

const associateSessionWithAccount = `-- name: AssociateSessionWithAccount :exec
INSERT INTO accounts_sessions (
  account_id, token, createdBy
) VALUES (
  ?1, ?2, ?3
)
`

type AssociateSessionWithAccountParams struct {
	AccountID int64
	Token     string
	Createdby string
}

func (q *Queries) AssociateSessionWithAccount(ctx context.Context, arg AssociateSessionWithAccountParams) error {
	_, err := q.db.ExecContext(ctx, associateSessionWithAccount, arg.AccountID, arg.Token, arg.Createdby)
	return err
}

const createAccountAndReturnId = `-- name: CreateAccountAndReturnId :one
INSERT INTO accounts (
  createdAt, updatedAt
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING id
`

func (q *Queries) CreateAccountAndReturnId(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, createAccountAndReturnId)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const findAccountByCallsign = `-- name: FindAccountByCallsign :one
SELECT accounts.id, accounts.name, accounts.createdat, accounts.updatedat, accounts.deletedat, accounts.kind
FROM accounts
JOIN accounts_callsigns ON accounts.id = accounts_callsigns.account_id
JOIN callsigns ON accounts_callsigns.callsign_id = callsigns.id
WHERE UPPER(callsigns.callsign) = UPPER(?1)
`

func (q *Queries) FindAccountByCallsign(ctx context.Context, upper string) (Account, error) {
	row := q.db.QueryRowContext(ctx, findAccountByCallsign, upper)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Deletedat,
		&i.Kind,
	)
	return i, err
}

const findAccountByEmail = `-- name: FindAccountByEmail :one
SELECT accounts.id, accounts.name, accounts.createdat, accounts.updatedat, accounts.deletedat, accounts.kind
FROM accounts
JOIN emails ON emails.account_id = accounts.id
WHERE emails.address = ?1
`

func (q *Queries) FindAccountByEmail(ctx context.Context, address string) (Account, error) {
	row := q.db.QueryRowContext(ctx, findAccountByEmail, address)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Deletedat,
		&i.Kind,
	)
	return i, err
}

const getAccount = `-- name: GetAccount :one
SELECT accounts.id, accounts.name, accounts.createdat, accounts.updatedat, accounts.deletedat, accounts.kind
FROM accounts
WHERE id = ?1
LIMIT 1
`

func (q *Queries) GetAccount(ctx context.Context, id int64) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccount, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Deletedat,
		&i.Kind,
	)
	return i, err
}

const updateAccount = `-- name: UpdateAccount :one
UPDATE accounts
SET updatedAt = CURRENT_TIMESTAMP,
    name = ?2
WHERE id = ?1
RETURNING id, name, createdat, updatedat, deletedat, kind
`

type UpdateAccountParams struct {
	ID   int64
	Name string
}

func (q *Queries) UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, updateAccount, arg.ID, arg.Name)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Deletedat,
		&i.Kind,
	)
	return i, err
}

const accounts = `-- name: accounts :many
SELECT id, name, createdat, updatedat, deletedat, kind FROM accounts
`

func (q *Queries) accounts(ctx context.Context) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, accounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Createdat,
			&i.Updatedat,
			&i.Deletedat,
			&i.Kind,
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
