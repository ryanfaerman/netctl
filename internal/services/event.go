package services

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/r3labs/sse/v2"
	"github.com/ryanfaerman/netctl/internal/dao"
	"github.com/ryanfaerman/netctl/internal/events"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/workgroup"
)

type event struct {
	Server *sse.Server

	subscribers map[string][]subscriber // event.Name: []subscriber
	ticker      *time.Ticker
	stopCh      chan bool
}

type subscriber struct {
	fn   func(context.Context, models.Event) error
	name string
}

var Event = &event{
	subscribers: make(map[string][]subscriber),
}

func (e *event) Create(ctx context.Context, stream string, evt any) error {
	d, err := json.Marshal(evt)
	if err != nil {
		global.log.Error("unable to marshal event", "error", err)
	}

	id, err := global.dao.CreateEvent(ctx, dao.CreateEventParams{
		StreamID:  stream,
		AccountID: Session.GetAccount(ctx).ID,
		EventType: fmt.Sprintf("%T", evt),
		EventData: string(d),
	})
	if err != nil {
		return err
	}

	j, err := events.Decode(fmt.Sprintf("%T", evt), d)
	if err != nil {
		return err
	}

	event := models.Event{
		ID:        id,
		At:        time.Now(),
		StreamID:  stream,
		AccountID: 1,
		Name:      fmt.Sprintf("%T", evt),
		Event:     j,
	}

	go e.Publish(context.WithoutCancel(ctx), event)

	return nil
}

func (e *event) Register(event any, fn func(context.Context, models.Event) error) {
	eventName := fmt.Sprintf("%T", event)
	funcName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()

	global.log.Debug("registering event subscriber", "event", eventName, "func", funcName)

	e.subscribers[eventName] = append(e.subscribers[eventName], subscriber{
		name: funcName,
		fn:   fn,
	})
}

func (e *event) Publish(ctx context.Context, event models.Event) error {
	l := global.log.With("task", "event-publish", "event", event.Name)

	if len(e.subscribers[event.Name]) == 0 {
		l.Debug("no subscribers for event")
		return nil
	}
	wg := workgroup.New(5) // TODO: make this configurable

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Minute) // TODO: make this configurable
	defer cancel()

	for _, sub := range e.subscribers[event.Name] {
		sub := sub
		wg.Go(func() error {
			l = l.With("subscriber", sub.name)

			if err := wg.Acquire(1); err != nil {
				l.Warn("error acquiring workgroup", "error", err)
				return err
			}

			l.Debug("creating recovery tombstone")

			recoveryID, err := global.dao.CreateEventRecovery(ctx, dao.CreateEventRecoveryParams{
				EventsID:     event.ID,
				RegisteredFn: sub.name,
				SessionToken: global.session.Token(ctx),
			})
			if err != nil {
				l.Error("error creating recovery tombstone", "error", err)
				return err
			}

			l.Debug("publishing event")
			if err := sub.fn(ctx, event); err != nil {
				l.Error("error publishing event", "error", err)
				return err
			}

			if err := global.dao.DeleteEventRecovery(ctx, recoveryID); err != nil {
				l.Error("error deleting recovered event", "error", err)
				return err
			}

			return nil
		})
	}
	return wg.Wait()
	// delete recovery tombstone
}

func (e *event) Recover(ctx context.Context) error {
	l := global.log.With("task", "event-recovery")
	recovereds, err := models.FindRecoverableEvents(ctx)
	if err != nil {
		l.Error("error finding recoverable events", "error", err)
		return err
	}

	if len(recovereds) == 0 {
		l.Debug("no recoverable events found")
		return nil
	}
	l.Debug("recovering events", "count", len(recovereds))

	wg := workgroup.New(5) // TODO: make this configurable
	for _, recovered := range recovereds {
		ctx, err := global.session.Load(ctx, recovered.SessionToken)
		if err != nil {
			// cannot load the session, so it must not exist,
			// thus... maybe we just delete these events? or log it?
			// really need to think of a good approach here, cause that info would get lost if they log out.
			// maybe need to have logout not delete the actual token, just the data within.
			return fmt.Errorf("unable to load session: %w", err)
		}
		for _, sub := range e.subscribers[recovered.Event.Name] {
			l = l.With("subscriber", sub.name, "event", recovered.Event.Name)
			if recovered.RegisteredFn != sub.name {
				l.Debug("skipping recovery func", "registered", recovered.RegisteredFn)
				continue
			}
			l.Debug("running recovery func")

			recovered := recovered
			sub := sub
			wg.Go(func() error {
				if err := wg.Acquire(1); err != nil {
					l.Warn("error acquiring workgroup", "error", err)
					return err
				}

				l.Debug("publishing recovered event", "name", sub.name)
				if err := sub.fn(ctx, recovered.Event); err != nil {
					l.Error("error publishing recovered event", "error", err)
					return err
				}

				l.Debug("deleting recovered event")
				if err := global.dao.DeleteEventRecovery(ctx, recovered.ID); err != nil {
					l.Error("error deleting recovered event", "error", err)
					return err
				}
				return nil
			})

		}
	}
	return wg.Wait()
}

func (e *event) StartRecoveryService(every time.Duration) {
	l := global.log.With("service", "event-recovery")

	e.stopCh = make(chan bool)
	e.ticker = time.NewTicker(every)
	go func() {
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		l.Info("recovery service", "lifecycle", "started")
		for {
			select {
			case <-e.ticker.C:
				l.Debug("recovery service", "lifecycle", "running")
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(every))
				err := e.Recover(ctx)
				cancel()
				if err != nil {
					l.Error("recovery service", "lifecycle", "finished", "error", err)
				} else {
					l.Debug("recovery service", "lifecycle", "finished")
				}
			case <-e.stopCh:
				cancel()
				close(e.stopCh)
				l.Info("recovery service", "lifecycle", "stopped")
				return
			}
		}
	}()
}

func (e *event) StopRecoveryService() {
	if e.ticker == nil {
		global.log.Info("recovery service", "service", "event-recover", "lifecycle", "stopped")
		return
	}
	global.log.Info("recovery service", "service", "event-recover", "lifecycle", "stopping")
	e.ticker.Stop()
	e.ticker = nil
	e.stopCh <- true
}
