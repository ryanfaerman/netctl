// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: emails.sql

package dao

import (
	"context"
)

const addPrimaryEmailForAccount = `-- name: AddPrimaryEmailForAccount :exec
INSERT INTO emails (
  createdAt, updatedAt, account_id, address, isPrimary, verifiedAt
)
VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?1, ?2, true, CURRENT_TIMESTAMP)
`

type AddPrimaryEmailForAccountParams struct {
	AccountID int64
	Address   string
}

func (q *Queries) AddPrimaryEmailForAccount(ctx context.Context, arg AddPrimaryEmailForAccountParams) error {
	_, err := q.db.ExecContext(ctx, addPrimaryEmailForAccount, arg.AccountID, arg.Address)
	return err
}

const addSecondaryEmailForAccount = `-- name: AddSecondaryEmailForAccount :exec
INSERT INTO emails (
  createdAt, updatedAt, account_id, address, isPrimary
)
VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?1, ?2, false)
`

type AddSecondaryEmailForAccountParams struct {
	AccountID int64
	Address   string
}

func (q *Queries) AddSecondaryEmailForAccount(ctx context.Context, arg AddSecondaryEmailForAccountParams) error {
	_, err := q.db.ExecContext(ctx, addSecondaryEmailForAccount, arg.AccountID, arg.Address)
	return err
}

const getEmailsForAccount = `-- name: GetEmailsForAccount :many
SELECT emails.id, emails.createdat, emails.updatedat, emails.account_id, emails.address, emails.isprimary, emails.ispublic, emails.isnotifiable, emails.verifiedat
FROM emails
WHERE emails.account_id = ?1
`

func (q *Queries) GetEmailsForAccount(ctx context.Context, accountID int64) ([]Email, error) {
	rows, err := q.db.QueryContext(ctx, getEmailsForAccount, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Email
	for rows.Next() {
		var i Email
		if err := rows.Scan(
			&i.ID,
			&i.Createdat,
			&i.Updatedat,
			&i.AccountID,
			&i.Address,
			&i.Isprimary,
			&i.Ispublic,
			&i.Isnotifiable,
			&i.Verifiedat,
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

const setEmailAsPrimary = `-- name: SetEmailAsPrimary :exec
UPDATE emails
SET isPrimary = true
WHERE id = ?1
`

func (q *Queries) SetEmailAsPrimary(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, setEmailAsPrimary, id)
	return err
}

const setEmailAsSecondary = `-- name: SetEmailAsSecondary :exec
UPDATE emails
SET isPrimary = false
WHERE id = ?1
`

func (q *Queries) SetEmailAsSecondary(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, setEmailAsSecondary, id)
	return err
}