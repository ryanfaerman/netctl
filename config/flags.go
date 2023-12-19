package config

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	dao "github.com/ryanfaerman/netctl/config/data"
)

func (c *config) Flags() ([]ConfigOption, error) {
	c.WaitForLoad()

	out := []ConfigOption{}

	opts, err := c.queries.Flags(context.Background())
	if err != nil {
		return out, fmt.Errorf("cannot get flags; %w", err)
	}

	for _, opt := range opts {
		val := "undefined"
		if opt.Value.Valid {
			if opt.Value.Bool {
				val = "true"
			} else {
				val = "false"
			}
		}
		out = append(out, ConfigOption{
			Uri:  c.unescapeUri(opt.Uri),
			Data: val,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Uri < out[j].Uri
	})

	return out, nil
}

func (c *config) DefineFlag(uri string, d ...bool) error {
	c.WaitForLoad()

	uri = c.escapeUri(uri)

	fallback := false
	if len(d) > 0 {
		fallback = d[0]
	}

	return c.queries.DefineFlag(context.Background(), dao.DefineFlagParams{
		Uri: uri,
		Value: sql.NullBool{
			Bool:  fallback,
			Valid: len(d) > 0,
		},
	})
}

func (c *config) Flag(uri string, d ...bool) (bool, error) {
	c.WaitForLoad()

	uri = c.escapeUri(uri)

	fallback := false
	if len(d) > 0 {
		fallback = d[0]
	}

	v, err := c.queries.GetFlag(context.Background(), uri)
	if err != nil {
		if err == sql.ErrNoRows {
			return fallback, nil
		}
		return fallback, err
	}
	if v.Valid {
		return v.Bool, nil
	}

	return fallback, nil
}

func (c *config) IsFlag(uri string) bool {
	c.WaitForLoad()

	uri = c.escapeUri(uri)

	_, err := c.queries.GetFlag(context.Background(), uri)
	return err == nil
}

func (c *config) SetFlag(uri string, v bool) error {
	c.WaitForLoad()

	uri = c.escapeUri(uri)

	return c.queries.SetFlag(context.Background(), dao.SetFlagParams{
		Uri:   uri,
		Value: sql.NullBool{Bool: v, Valid: true},
	})
}
