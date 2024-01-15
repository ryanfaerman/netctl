-- create_nets
-- +goose Up
-- +goose StatementBegin
CREATE TABLE nets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted TIMESTAMP
);

CREATE TABLE net_sessions (
  id INTEGER PRIMARY KEY,
  net_id INTEGER NOT NULL,
  stream_id TEXT NOT NULL,
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(net_id) REFERENCES nets(id)
);

CREATE TABLE events (
  id INTEGER PRIMARY KEY,
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  stream_id TEXT NOT NULL,
  account_id INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  event_data BLOB NOT NULL,

  FOREIGN KEY(account_id) REFERENCES accounts(id)
);

CREATE INDEX idx_stream_id ON events(stream_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE nets;
DROP TABLE net_events;
DROP TABLE events;
DROP INDEX idx_stream_id;
-- +goose StatementEnd
