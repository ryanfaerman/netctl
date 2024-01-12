-- name: Users :many
SELECT * FROM users;

-- name: GetUser :one
SELECT users.*
FROM users
WHERE id = ?1
LIMIT 1;

-- name: FindUserByEmail :one
SELECT users.*
FROM users
JOIN emails ON emails.user_id = users.id
WHERE emails.address = ?1;

-- name: FindUserByCallsign :one
SELECT users.*
FROM users
JOIN users_callsigns ON users.id = users_callsigns.user_id
JOIN callsigns ON users_callsigns.callsign_id = callsigns.id
WHERE callsigns.callsign = ?1;

-- name: GetCallsignsForUser :many
SELECT callsigns.*
FROM callsigns
JOIN users_callsigns ON callsigns.id = users_callsigns.callsign_id
WHERE users_callsigns.user_id = ?1;

-- name: GetEmailsForUser :many
SELECT emails.*
FROM emails
WHERE emails.user_id = ?1;

-- name: CreateUserAndReturnId :one
INSERT INTO users (
  createdAt, updatedAt
) VALUES (
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING id;

-- name: AddPrimaryEmailForUser :exec
INSERT INTO emails (
  createdAt, updatedAt, user_id, address, isPrimary, verifiedAt
)
VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?1, ?2, true, CURRENT_TIMESTAMP);

-- name: AddSecondaryEmailForUser :exec
INSERT INTO emails (
  createdAt, updatedAt, user_id, address, isPrimary
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

-- name: AssociateSessionWithUser :exec
INSERT INTO users_sessions (
  user_id, token, createdBy
) VALUES (
  ?1, ?2, ?3
);
