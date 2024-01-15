package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/ryanfaerman/netctl/internal/dao"
)

type event struct{}

var Event event

func (event) Create(ctx context.Context, stream string, e any) error {
	var b bytes.Buffer
	var p any
	p = &e
	if err := gob.NewEncoder(&b).Encode(p); err != nil {
		return err
	}

	return global.dao.CreateEvent(ctx, dao.CreateEventParams{
		StreamID:  stream,
		AccountID: 1,
		EventType: fmt.Sprintf("%T", e),
		EventData: b.Bytes(),
	})

	return nil
}
