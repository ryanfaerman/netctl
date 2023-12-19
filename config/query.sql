/* name: SetConfig :exec */
INSERT INTO configs (uri, data)
VALUES (?1, ?2)
ON CONFLICT(uri) DO UPDATE
SET data = ?2;

/* name: GetConfig :one */
SELECT data FROM configs
WHERE uri = ?;

/* name: Configs :many */
SELECT * FROM configs;

/* name: UnsetConfig :exec */
DELETE FROM configs WHERE uri=?;

-- name: DefineConfig :exec
INSERT INTO configs (uri, data)
VALUES (?1, ?2)
ON CONFLICT(uri) DO NOTHING;


--- Flags

-- name: SetFlag :exec
INSERT INTO flags (uri, value)
VALUES (?1, ?2)
ON CONFLICT(uri) DO UPDATE
SET value = ?2;

-- name: DefineFlag :exec
INSERT INTO flags (uri, value)
VALUES (?1, ?2)
ON CONFLICT(uri) DO NOTHING;

-- name: SetFlagDefault :exec
INSERT INTO flags (uri, value)
VALUES (?1, ?2)
ON CONFLICT DO NOTHING;

-- name: GetFlag :one
SELECT value FROM flags
WHERE uri = ?;

-- name: Flags :many
SELECT * FROM flags;
