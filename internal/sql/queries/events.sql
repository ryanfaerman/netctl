-- name: CreateEvent :exec
INSERT INTO events (
  created, stream_id, account_id, event_type, event_data
)
VALUES (
  CURRENT_TIMESTAMP, ?1, ?2, ?3, ?4
);

-- name: GetEventsForStreams :many
SELECT * FROM events
WHERE stream_id IN (sqlc.slice('stream_ids'));

-- name: GetEventsForStream :many
SELECT * FROM events
WHERE stream_id = ?1;
