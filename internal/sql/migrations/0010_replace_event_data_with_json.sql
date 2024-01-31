-- replace_event_data_with_json
-- +goose Up
-- +goose StatementBegin

ALTER TABLE events
DROP COLUMN event_data;

ALTER TABLE events
RENAME COLUMN event_data_json TO event_data;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

