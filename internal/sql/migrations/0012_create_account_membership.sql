-- create_account_membership
-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts
ADD COLUMN slug TEXT NOT NULL DEFAULT '';

CREATE TABLE memberships (
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL,
    member_of INTEGER NOT NULL,
    role_id INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (account_id, member_of),
    FOREIGN KEY (member_of) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (role_id) REFERENCES roles(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE roles (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    account_id INTEGER NOT NULL,
    permissions INTEGER NOT NULL DEFAULT 0,
    ranking INTEGER NOT NULL DEFAULT 0, 
    FOREIGN KEY (account_id) REFERENCES accounts(id) DEFERRABLE INITIALLY DEFERRED
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE memberships;
DROP TABLE roles;

ALTER TABLE accounts
DROP COLUMN slug;
-- +goose StatementEnd

