-- name: CreateCallsignAndReturnId :one
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
RETURNING id;

-- name: AssociateCallsignWithAccount :exec
INSERT INTO accounts_callsigns (
  account_id, callsign_id
) VALUES (
  ?1, ?2
);

-- name: FindCallsignsForAccount :many
SELECT callsigns.*
FROM callsigns
JOIN accounts_callsigns ON callsigns.id = accounts_callsigns.callsign_id
WHERE accounts_callsigns.account_id = ?1;

-- name: FindCallsign :one
SELECT *
FROM callsigns
WHERE callsigns.callsign = ?1;
