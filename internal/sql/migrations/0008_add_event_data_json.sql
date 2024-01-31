-- add_event_data_json
-- +goose Up
-- +goose StatementBegin

ALTER TABLE events
ADD COLUMN event_data_json BLOB NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE events
DROP COLUMN event_data_json;
-- +goose StatementEnd

