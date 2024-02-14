-- name: AddVerifiedEmailForAccount :exec
INSERT INTO emails (
  createdAt, account_id, address, verifiedAt
)
VALUES (CURRENT_TIMESTAMP, ?1, ?2, CURRENT_TIMESTAMP);


-- name: GetEmailsForAccount :many
SELECT emails.*
FROM emails
WHERE emails.account_id = ?1;

-- name: AddEmailForAccount :one
INSERT INTO emails (
  createdAt, updatedAt, account_id, address
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?1, ?2
) RETURNING id;

-- name: SetEmailVerified :exec
UPDATE emails
SET verifiedAt = CURRENT_TIMESTAMP
WHERE id = ?1;

-- name: GetEmail :one
SELECT *
FROM emails
WHERE id = ?1;

-- name: DeleteEmail :exec
DELETE FROM emails
WHERE id = ?1;
