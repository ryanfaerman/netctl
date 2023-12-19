package hook

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/charmbracelet/log"
	"github.com/ryanfaerman/netctl/workgroup"
)

var (
	Logger = log.Default()
)

var id uint64

func nextID() uint64 {
	return atomic.AddUint64(&id, 1)
}

// Hook is a type that can be used to register and dispatch hooks
type Hook[T any] struct {
	name  string
	fns   map[uint64]Listener[T]
	names map[uint64]string
	limit int64

	tombstones map[uint64]uint64

	fnsMtx sync.RWMutex
	tmbMtx sync.RWMutex
}

// New creates a new Hook with the given name. The default concurrency limit is
// 1. Use WithLimit to change this.
func New[T any](name string) *Hook[T] {
	return &Hook[T]{
		name:       name,
		fns:        make(map[uint64]Listener[T]),
		names:      make(map[uint64]string),
		tombstones: make(map[uint64]uint64),
		limit:      1,
	}
}

// WithLimit returns a new Hook with the given concurrency limit. This is
// intended for use when creating a Hook. A 0 limit is invalid and is redefined
// as a limit of 1.
func (h *Hook[T]) WithLimit(limit int) *Hook[T] {
	if limit == 0 {
		limit = 1
	}
	h.SetLimit(limit)
	return h
}

// SetLimit of the Hook. This can be used to change the concurrency limit after
// the initial creation. A negative limit means that the concurrent is
// unlimited.
func (h *Hook[T]) SetLimit(n int) {
	Logger.Debug("setting limit", "hook", h.Name(), "limit", n)

	h.limit = int64(n)

}

// Name of the Hook
func (h *Hook[T]) Name() string {
	return h.name
}

// ListenerCount returns the number of listeners registered with the Hook.
func (h *Hook[T]) ListenerCount() int {
	h.fnsMtx.RLock()
	defer h.fnsMtx.RUnlock()
	return len(h.fns)
}

// Register a new listener for the Hook.
func (h *Hook[T]) Register(fn Listener[T]) {
	h.fnsMtx.Lock()
	defer h.fnsMtx.Unlock()

	id := nextID()

	if _, file, lineno, ok := runtime.Caller(1); ok {
		caller := fmt.Sprintf("%s:%d", file, lineno)
		h.names[id] = caller
	} else {
		h.names[id] = "unknown"
	}

	h.fns[id] = fn

	Logger.Debug("registered", "hook", h.Name(), "index", id, "caller", h.names[id])
}

// Unregister a listener from the hook by the given function index. This is
// intended to called from within a listener.
func (h *Hook[T]) unregister(fnIndex uint64) {
	Logger.Debug("unregistering",
		"hook", h.Name(),
		"index", fnIndex,
	)

	h.tmbMtx.Lock()
	defer h.tmbMtx.Unlock()
	h.tombstones[fnIndex] = fnIndex
}

func (h *Hook[T]) isTombStoned(id uint64) bool {
	h.tmbMtx.RLock()
	_, ok := h.tombstones[id]
	h.tmbMtx.RUnlock()
	return ok
}

// Dispatch the payload to all registered listeners. The context is used to
// manage timeouts and whatever else contexts can do. Listeners are invoked
// concurently according to the Hook's limit. If the context is Done() the
// dispatch ends and handlers will no longer be invoked.
func (h *Hook[T]) Dispatch(ctx context.Context, payload T) error {
	wg := workgroup.WithContext(ctx, h.limit)

	Logger.Debug("dispatching", "hook", h.name, "limit", h.limit)

	h.fnsMtx.RLock()
	defer h.fnsMtx.RUnlock()

	for k, fn := range h.fns {
		fn := fn
		k := k

		if h.isTombStoned(k) {
			continue
		}

		wg.Go(func() error {
			if err := wg.Acquire(1); err != nil {
				return err
			}
			defer wg.Release(1)

			e := newEvent(payload, h)
			e.Context = ctx
			e.id = k

			callback := func() <-chan struct{} {
				ch := make(chan struct{})
				go func() {
					defer close(ch)
					fn(e)
				}()
				return ch

			}

			select {
			case <-ctx.Done():
				Logger.Error("dispatch failed, context cancelled",
					"hook", h.Name(),
					"source", h.names[k],
				)
				return fmt.Errorf(
					"dispatch failed, context cancelled; source: %s",
					h.names[k],
				)

			case <-callback():
			}

			return e.Error
		})

	}

	err := wg.Wait()

	h.tmbMtx.Lock()
	if len(h.tombstones) > 0 {
		Logger.Debug("clearing tombstones", "hook", h.Name())
		for id := range h.tombstones {
			delete(h.fns, id)
			delete(h.names, id)
			delete(h.tombstones, id)
		}
	}
	h.tmbMtx.Unlock()

	return err

}
