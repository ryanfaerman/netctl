-- name: accounts :many
SELECT * FROM accounts;

-- name: GetAccount :one
SELECT accounts.*
FROM accounts
WHERE id = ?1
LIMIT 1;

-- name: UpdateAccount :one
UPDATE accounts
SET updatedAt = CURRENT_TIMESTAMP,
    name = ?2,
    about = ?3,
    kind = ?4
WHERE id = ?1
RETURNING *;

-- name: FindAccountByEmail :one
SELECT accounts.*
FROM accounts
JOIN emails ON emails.account_id = accounts.id
WHERE emails.address = ?1;

-- name: FindAccountByCallsign :one
SELECT accounts.*
FROM accounts
JOIN accounts_callsigns ON accounts.id = accounts_callsigns.account_id
JOIN callsigns ON accounts_callsigns.callsign_id = callsigns.id
WHERE UPPER(callsigns.callsign) = UPPER(?1);

-- name: CreateAccountAndReturnId :one
INSERT INTO accounts (
  createdAt, updatedAt
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING id;

-- name: AssociateSessionWithAccount :exec
INSERT INTO accounts_sessions (
  account_id, token, createdBy
) VALUES (
  ?1, ?2, ?3
);
