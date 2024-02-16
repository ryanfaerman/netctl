-- add_token_to_event_recovery
-- +goose Up
-- +goose StatementBegin
ALTER TABLE events_recovery 
ADD COLUMN session_token text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events_recovery 
DROP COLUMN session_token;
-- +goose StatementEnd

