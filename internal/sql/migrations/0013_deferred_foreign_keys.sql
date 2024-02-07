-- deferred_foreign_keys
-- +goose Up
-- +goose StatementBegin
CREATE TABLE temp_accounts_sessions (
  account_id integer NOT NULL,
  token text NOT NULL,

  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  createdBy text NOT NULL DEFAULT 'system',

  PRIMARY KEY (account_id, token),

  FOREIGN KEY (account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED,
  FOREIGN KEY (token) REFERENCES sessions(token) DEFERRABLE INITIALLY DEFERRED
);
INSERT INTO temp_accounts_sessions SELECT * FROM accounts_sessions;
DROP TABLE accounts_sessions;
ALTER TABLE temp_accounts_sessions RENAME TO accounts_sessions;


CREATE TABLE temp_emails (
  id integer PRIMARY KEY,

  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

  account_id integer NOT NULL,

  address text NOT NULL,
  isPrimary boolean NOT NULL DEFAULT false,
  isPublic boolean NOT NULL DEFAULT false,
  isNotifiable boolean NOT NULL DEFAULT true,
  verifiedAt timestamp,

  FOREIGN KEY (account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED,
  UNIQUE (account_id, isPrimary),
  UNIQUE (address)
);

INSERT INTO temp_emails SELECT * FROM emails;
DROP TABLE emails;
ALTER TABLE temp_emails RENAME TO emails;


CREATE TABLE temp_accounts_callsigns (
  account_id integer NOT NULL,
  callsign_id integer NOT NULL,

  PRIMARY KEY (account_id, callsign_id),
  UNIQUE (callsign_id),
  FOREIGN KEY (account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED,
  FOREIGN KEY (callsign_id) REFERENCES callsigns(id) DEFERRABLE INITIALLY DEFERRED
);
INSERT INTO temp_accounts_callsigns SELECT * FROM accounts_callsigns;
DROP TABLE accounts_callsigns;
ALTER TABLE temp_accounts_callsigns RENAME TO accounts_callsigns;


CREATE TABLE temp_events_recovery (
  id INTEGER PRIMARY KEY,
  events_id INTEGER NOT NULL,
  registered_fn TEXT NOT NULL DEFAULT '',
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(events_id) REFERENCES events(id) DEFERRABLE INITIALLY DEFERRED
);

INSERT INTO temp_events_recovery SELECT * FROM events_recovery;
DROP TABLE events_recovery;
ALTER TABLE temp_events_recovery RENAME TO events_recovery;


CREATE TABLE temp_net_sessions (
  id INTEGER PRIMARY KEY,
  net_id INTEGER NOT NULL,
  stream_id TEXT NOT NULL,
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY(net_id) REFERENCES nets(id) DEFERRABLE INITIALLY DEFERRED
);
INSERT INTO temp_net_sessions SELECT * FROM net_sessions;
DROP TABLE net_sessions;
ALTER TABLE temp_net_sessions RENAME TO net_sessions;


CREATE TABLE temp_events (
  id INTEGER PRIMARY KEY,
  created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  stream_id TEXT NOT NULL,
  account_id INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  event_data BLOB NOT NULL,

  FOREIGN KEY(account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED
);
INSERT INTO temp_events SELECT * FROM events;
DROP TABLE events;
ALTER TABLE temp_events RENAME TO events;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

