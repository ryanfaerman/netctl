-- +goose Up
-- +goose StatementBegin
CREATE TABLE nets (
  id integer PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deletedAt timestamp,

  -- attributes
  name text,
  tx_frequency real,
  rx_frequency real,
  tone real,
  preamble text,
  postamble text
);

CREATE TABLE net_sessions (
  id integer PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deletedAt timestamp,

  -- attributes
  startedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  closedAt timestamp,

  -- relations
  net_id integer REFERENCES nets(id) NOT NULL
);

CREATE TABLE checkins (
  id integer PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deletedAt timestamp,

  -- attributes
  callsign text REFERENCES callsigns(callsign) NOT NULL,
  short_on_time boolean NOT NULL DEFAULT false,
  traffic boolean NOT NULL DEFAULT false,
  acknowledged boolean NOT NULL DEFAULT false,
  announcements boolean NOT NULL DEFAULT false,

  -- relations
  net_session_id integer REFERENCES net_sessions(id) NOT NULL
);

CREATE TABLE callsigns (
  callsign text NOT NULL PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

  -- attributes
  name text
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE nets;
DROP TABLE net_sessions;
DROP TABLE checkins;
DROP TABLE callsigns;
-- +goose StatementEnd
