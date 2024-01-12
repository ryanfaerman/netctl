-- create_users
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
	id integer PRIMARY KEY,

  name text NOT NULL DEFAULT '',

  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deletedAt timestamp
);

CREATE TABLE users_sessions (
  user_id integer NOT NULL,
  token text NOT NULL,

  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  createdBy text NOT NULL DEFAULT 'system',

  PRIMARY KEY (user_id, token),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (token) REFERENCES sessions(token)
);

CREATE TABLE emails (
  id integer PRIMARY KEY,

  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

  user_id integer NOT NULL,

  address text NOT NULL,
  isPrimary boolean NOT NULL DEFAULT false,
  isPublic boolean NOT NULL DEFAULT false,
  isNotifiable boolean NOT NULL DEFAULT true,
  verifiedAt timestamp,

  FOREIGN KEY (user_id) REFERENCES users(id),
  UNIQUE (user_id, isPrimary),
  UNIQUE (address)

);

CREATE TABLE callsigns (
  id integer PRIMARY KEY,

  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

  callsign text NOT NULL,
  class integer NOT NULL DEFAULT 0,
  expires timestamp,
  status integer NOT NULL DEFAULT 0,
  grid text,
  latitude real,
  longitude real,
  firstName text,
  middleName text,
  lastName text,
  suffix text,
  address text,
  city text,
  state text,
  zip text,
  country text

);

CREATE TABLE users_callsigns (
  user_id integer NOT NULL,
  callsign_id integer NOT NULL,

  PRIMARY KEY (user_id, callsign_id),
  UNIQUE (callsign_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (callsign_id) REFERENCES callsigns(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE emails;
DROP TABLE callsigns;
DROP TABLE users_callsigns;
-- +goose StatementEnd
