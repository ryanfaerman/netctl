package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
)

func init() {
	fmt.Println("hello world")
	AddMigration(Up0009, Down0009)
}

func Up0009(ctx context.Context, tx *sql.Tx) error {
	l := Log.With("migration", "0009", "direction", "up")

	l.Debug("starting migration")

	rows, err := tx.QueryContext(ctx, `select id, event_data from events`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var decoder *gob.Decoder

	for rows.Next() {
		var (
			id   int64
			data []byte
		)
		if err := rows.Scan(&id, &data); err != nil {
			return err
		}

		l := l.With("event.id", id)

		l.Debug("decoding event data")
		decoder = gob.NewDecoder(bytes.NewReader(data))
		var p any
		if err := decoder.Decode(&p); err != nil {
			l.Error("unable to decode event", "error", err)
			return err
		}

		l.Debug("marshaling event data to json")
		jsonData, err := json.Marshal(p)
		if err != nil {
			return err
		}

		l.Debug("updating event data json")
		if _, err := tx.ExecContext(
			ctx,
			`update events set event_data_json = ? where id = ?`,
			jsonData, id,
		); err != nil {
			return err
		}
		l.Debug("updated event data json")

	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func Down0009(ctx context.Context, tx *sql.Tx) error {
	return errors.New("cannot downgrade")
}
