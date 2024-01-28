-- add_stream_id_to_account
-- +goose Up
-- +goose StatementBegin
--
ALTER TABLE nets
ADD COLUMN stream_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_nets_stream_id ON nets(stream_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE nets
DROP COLUMN stream_id;

DROP INDEX idx_nets_stream_id;
-- +goose StatementEnd

