-- name: AddPrimaryEmailForAccount :exec
INSERT INTO emails (
  createdAt, updatedAt, account_id, address, isPrimary, verifiedAt
)
VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?1, ?2, true, CURRENT_TIMESTAMP);

-- name: AddSecondaryEmailForAccount :exec
INSERT INTO emails (
  createdAt, updatedAt, account_id, address, isPrimary
)
VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?1, ?2, false);

-- name: SetEmailAsPrimary :exec
UPDATE emails
SET isPrimary = true
WHERE id = ?1;

-- name: SetEmailAsSecondary :exec
UPDATE emails
SET isPrimary = false
WHERE id = ?1;

-- name: GetEmailsForAccount :many
SELECT emails.*
FROM emails
WHERE emails.account_id = ?1;
