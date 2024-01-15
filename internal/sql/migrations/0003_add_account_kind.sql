-- create_accounts
-- +goose Up
-- +goose StatementBegin

ALTER TABLE accounts
ADD COLUMN kind INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
DROP COLUMN kind;
-- +goose StatementEnd
