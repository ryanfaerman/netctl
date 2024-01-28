-- add_event_recovery.sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE events_recovery (
  id INTEGER PRIMARY KEY,
  events_id INTEGER NOT NULL,
  registered_fn TEXT NOT NULL DEFAULT '',
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(events_id) REFERENCES events(id)
)


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events_recovery;
-- +goose StatementEnd

