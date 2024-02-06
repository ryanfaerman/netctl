-- add_user_settings
-- +goose Up
-- +goose StatementBegin
ALTER TABLE accounts
ADD COLUMN settings BLOB NOT NULL DEFAULT '{}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
DROP COLUMN settings;
-- +goose StatementEnd

