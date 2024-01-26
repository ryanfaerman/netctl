-- name: CreateNetSessionAndReturnId :one
INSERT INTO net_sessions (
  net_id, stream_id, created
)VALUES (
  ?1, ?2, CURRENT_TIMESTAMP
)
RETURNING id;

-- name: GetNetSessions :many
SELECT * FROM net_sessions WHERE net_id = ?1;

-- name: GetNetSessionEvents :many
SELECT events.*
FROM events
JOIN net_sessions ON events.stream_id = net_sessions.stream_id
WHERE net_sessions.net_id = ?1;

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

-- name: GetNets :many
SELECT * FROM nets;

-- name: GetNetForSession :one
SELECT nets.*, net_sessions.created AS session_created
FROM nets 
JOIN net_sessions ON net_sessions.net_id = nets.id
WHERE net_sessions.stream_id = ?1;
