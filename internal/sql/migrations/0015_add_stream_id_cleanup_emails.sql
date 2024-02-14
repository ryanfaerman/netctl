-- add_stream_id_cleanup_emails
-- +goose Up
-- +goose StatementBegin
CREATE TABLE temp_emails (
  id integer PRIMARY KEY,
  createdAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  account_id integer NOT NULL,
  address text NOT NULL,
  verifiedAt timestamp,
  FOREIGN KEY (account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED,
  UNIQUE (address)
);

INSERT INTO temp_emails (id, createdAt, updatedAt, account_id, address, verifiedAt)
SELECT id, createdAt, updatedAt, account_id, address, verifiedAt FROM emails;

DROP TABLE emails;
ALTER TABLE temp_emails RENAME TO emails;


ALTER TABLE accounts
ADD COLUMN stream_id text NOT NULL DEFAULT '';
UPDATE accounts SET stream_id = slug;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

