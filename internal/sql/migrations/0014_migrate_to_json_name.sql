-- migrate_to_json_name
-- +goose Up
-- +goose StatementBegin
UPDATE accounts SET settings = json_insert(settings, '$.profile.name', name);
UPDATE accounts SET settings = json_insert(settings, '$.profile.about', about);
ALTER TABLE accounts DROP COLUMN name;
ALTER TABLE accounts DROP COLUMN about;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE accounts ADD COLUMN name text NOT NULL default '';
ALTER TABLE accounts ADD COLUMN about text NOT NULL default '';
UPDATE accounts SET name = json_extract(settings, '$.profile.name');
UPDATE accounts SET about = json_extract(settings, '$.profile.about');

-- +goose StatementEnd

