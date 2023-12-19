package config

import (
	"context"
	"fmt"
	"sort"
)

type ConfigOption struct {
	Uri  string
	Data string
}

func (c *config) GetAll() ([]ConfigOption, error) {
	c.WaitForLoad()

	out := []ConfigOption{}

	opts, err := c.queries.Configs(context.Background())
	if err != nil {
		return out, fmt.Errorf("cannot get configs; %w", err)
	}

	for _, opt := range opts {
		val := "undefined"
		if opt.Data.Valid {
			val = opt.Data.String
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

func (c *config) Unset(uri string) error {
	return c.queries.UnsetConfig(context.Background(), uri)
}
