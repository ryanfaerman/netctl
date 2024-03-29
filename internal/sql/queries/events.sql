-- name: CreateEvent :one
INSERT INTO events (
  created, stream_id, account_id, event_type, event_data
)
VALUES (
  CURRENT_TIMESTAMP, ?1, ?2, ?3, ?4
) RETURNING id;

-- name: GetEventsForStreams :many
SELECT * FROM events
WHERE stream_id IN (sqlc.slice('stream_ids'));

-- name: GetEventsForStream :many
SELECT * FROM events
WHERE stream_id = ?1;

-- name: GetEvents :many
SELECT * FROM events 
where id IN (sqlc.slice('ids'));

-- name: GetRecoverableEvents :many
SELECT 
  events_recovery.id as recovery_id,
  events_recovery.registered_fn as registered_fn,
  events_recovery.session_token as session_token,
  events.*
FROM events_recovery
JOIN events ON events.id = events_recovery.events_id;

-- name: CreateEventRecovery :one
INSERT INTO events_recovery (
  events_id,
  registered_fn,
  session_token
) VALUES (
  ?1, ?2, ?3
) RETURNING id;

-- name: DeleteEventRecovery :exec
DELETE FROM events_recovery WHERE id = ?1;

-- name: GetEventsForCallsign :many
SELECT *
FROM events
WHERE event_type = ?1
AND json_extract(event_data, '$.callsign') = @callsign;
