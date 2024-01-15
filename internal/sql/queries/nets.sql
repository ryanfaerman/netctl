-- name: CreateNetEvent :exec
INSERT INTO net_events (
  created,
  net_id,
  session_id,
  account_id,
  event_type,
  event_data
) VALUES (
  CURRENT_TIMESTAMP,
  ?1,
  ?2,
  ?3,
  ?4,
  ?5
);

-- name: GetNetEvents :many
SELECT * FROM net_events WHERE net_id = ?1;

-- name: CreateNetAndReturnId :one
INSERT INTO nets (
  name,
  created,
  updated
) VALUES (
  ?1,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
RETURNING id;

-- name: GetNet :one
SELECT * FROM nets WHERE id = ?1;
