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

CREATE TABLE net_events (
  id INTEGER PRIMARY KEY,
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  net_id INTEGER NOT NULL,
  session_id TEXT NOT NULL,
  account_id INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  event_data BLOB NOT NULL,

  FOREIGN KEY(net_id) REFERENCES nets(id),
  FOREIGN KEY(account_id) REFERENCES accounts(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE nets;
DROP TABLE net_events;
-- +goose StatementEnd
