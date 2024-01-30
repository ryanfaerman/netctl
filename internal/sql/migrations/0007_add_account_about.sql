-- add-account-about
-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts
ADD COLUMN about TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
DROP COLUMN about;
-- +goose StatementEnd

