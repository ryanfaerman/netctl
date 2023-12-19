-- +goose Up
-- +goose StatementBegin
CREATE TABLE flags (
  uri text PRIMARY KEY,
  value boolean
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE flags;
-- +goose StatementEnd
