// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: events.sql

package dao

import (
	"context"
	"strings"
	"time"
)

const createEvent = `-- name: CreateEvent :one
INSERT INTO events (
  created, stream_id, account_id, event_type, event_data
)
VALUES (
  CURRENT_TIMESTAMP, ?1, ?2, ?3, ?4
) RETURNING id
`

type CreateEventParams struct {
	StreamID  string
	AccountID int64
	EventType string
	EventData []byte
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createEvent,
		arg.StreamID,
		arg.AccountID,
		arg.EventType,
		arg.EventData,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const createEventRecovery = `-- name: CreateEventRecovery :one
INSERT INTO events_recovery (
  events_id,
  registered_fn
) VALUES (
  ?1, ?2
) RETURNING id
`

type CreateEventRecoveryParams struct {
	EventsID     int64
	RegisteredFn string
}

func (q *Queries) CreateEventRecovery(ctx context.Context, arg CreateEventRecoveryParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, createEventRecovery, arg.EventsID, arg.RegisteredFn)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const deleteEventRecovery = `-- name: DeleteEventRecovery :exec
DELETE FROM events_recovery WHERE id = ?1
`

func (q *Queries) DeleteEventRecovery(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteEventRecovery, id)
	return err
}

const getEvents = `-- name: GetEvents :many
SELECT id, created, stream_id, account_id, event_type, event_data FROM events 
where id IN (/*SLICE:ids*/?)
`

func (q *Queries) GetEvents(ctx context.Context, ids []int64) ([]Event, error) {
	query := getEvents
	var queryParams []interface{}
	if len(ids) > 0 {
		for _, v := range ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Created,
			&i.StreamID,
			&i.AccountID,
			&i.EventType,
			&i.EventData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEventsForCallsign = `-- name: GetEventsForCallsign :many
SELECT id, created, stream_id, account_id, event_type, event_data
FROM events
WHERE event_type = ?1 
AND json_extract(event_data, '$.Callsign') = ?2
`

type GetEventsForCallsignParams struct {
	EventType string
	Callsign  []byte
}

func (q *Queries) GetEventsForCallsign(ctx context.Context, arg GetEventsForCallsignParams) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, getEventsForCallsign, arg.EventType, arg.Callsign)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Created,
			&i.StreamID,
			&i.AccountID,
			&i.EventType,
			&i.EventData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEventsForStream = `-- name: GetEventsForStream :many
SELECT id, created, stream_id, account_id, event_type, event_data FROM events
WHERE stream_id = ?1
`

func (q *Queries) GetEventsForStream(ctx context.Context, streamID string) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, getEventsForStream, streamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Created,
			&i.StreamID,
			&i.AccountID,
			&i.EventType,
			&i.EventData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEventsForStreams = `-- name: GetEventsForStreams :many
SELECT id, created, stream_id, account_id, event_type, event_data FROM events
WHERE stream_id IN (/*SLICE:stream_ids*/?)
`

func (q *Queries) GetEventsForStreams(ctx context.Context, streamIds []string) ([]Event, error) {
	query := getEventsForStreams
	var queryParams []interface{}
	if len(streamIds) > 0 {
		for _, v := range streamIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:stream_ids*/?", strings.Repeat(",?", len(streamIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:stream_ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.Created,
			&i.StreamID,
			&i.AccountID,
			&i.EventType,
			&i.EventData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRecoverableEvents = `-- name: GetRecoverableEvents :many
SELECT 
  events_recovery.id as recovery_id,
  events_recovery.registered_fn as registered_fn,
  events.id, events.created, events.stream_id, events.account_id, events.event_type, events.event_data
FROM events_recovery
JOIN events ON events.id = events_recovery.events_id
`

type GetRecoverableEventsRow struct {
	RecoveryID   int64
	RegisteredFn string
	ID           int64
	Created      time.Time
	StreamID     string
	AccountID    int64
	EventType    string
	EventData    []byte
}

func (q *Queries) GetRecoverableEvents(ctx context.Context) ([]GetRecoverableEventsRow, error) {
	rows, err := q.db.QueryContext(ctx, getRecoverableEvents)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRecoverableEventsRow
	for rows.Next() {
		var i GetRecoverableEventsRow
		if err := rows.Scan(
			&i.RecoveryID,
			&i.RegisteredFn,
			&i.ID,
			&i.Created,
			&i.StreamID,
			&i.AccountID,
			&i.EventType,
			&i.EventData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
