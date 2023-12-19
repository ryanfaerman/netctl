-- +goose Up
-- +goose StatementBegin
CREATE TABLE configs (
  uri text PRIMARY KEY,
  data text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE configs;
-- +goose StatementEnd
