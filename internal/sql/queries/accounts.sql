-- name: accounts :many
SELECT * FROM accounts;

-- name: GetAccount :one
SELECT accounts.*
FROM accounts
WHERE id = ?1
LIMIT 1;

-- name: GetAccountBySlug :one
SELECT accounts.*
FROM accounts
WHERE UPPER(slug) = UPPER(@slug)
LIMIT 1;


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

-- name: CreateAccount :one
INSERT INTO accounts (
  kind, slug
) VALUES (
  ?1, ?2
)
RETURNING id;

-- name: AssociateSessionWithAccount :exec
INSERT INTO accounts_sessions (
  account_id, token, createdBy
) VALUES (
  ?1, ?2, ?3
);

-- name: GetAccountSetting :one
SELECT json_extract(settings, @jsonpath)
FROM accounts
WHERE id = ?1;

-- name: SetAccountSetting :exec
UPDATE accounts
SET settings = json_set(settings, @jsonpath, @jsonvalue)
WHERE id = ?1;

-- name: UpdateAccountSettings :exec
UPDATE accounts
SET settings=json(@settings)
WHERE id = ?1;

-- name: CheckSlugAvailability :one
SELECT COUNT(*) as count FROM accounts WHERE slug = ?1;
