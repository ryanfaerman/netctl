package workgroup

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// WorkGroup marries a sync/semaphore with a sync/errgroup. It provides a way
// to both bound concurrent processing and bubble up errors from the concurrent
// processes.
type WorkGroup struct {
	ctx context.Context
	g   *errgroup.Group
	sem *semaphore.Weighted
}

// New creates a WorkGroup that has the given limit processes. If any
// process returns an error, the other processes *are not* halted.
func New(limit int64) *WorkGroup {
	wg := &WorkGroup{ctx: context.Background()}
	wg.g = new(errgroup.Group)
	wg.sem = semaphore.NewWeighted(limit)

	return wg
}

// WithContext creates a WorkGroup that limits concurrent
// processing and ties them together into a single group of work. Should any
// process return an error, or if the given context is done, all processes are
// halted.
func WithContext(ctx context.Context, limit int64) *WorkGroup {
	wg := &WorkGroup{}
	wg.g, wg.ctx = errgroup.WithContext(ctx)
	wg.sem = semaphore.NewWeighted(limit)

	return wg
}

func (wg *WorkGroup) Acquire(n int64) error { return wg.sem.Acquire(wg.ctx, n) }
func (wg *WorkGroup) AcquireWithContext(ctx context.Context, n int64) error {
	return wg.sem.Acquire(ctx, n)
}
func (wg *WorkGroup) TryAcquire(n int64) bool { return wg.sem.TryAcquire(n) }
func (wg *WorkGroup) Release(n int64)         { wg.sem.Release(n) }
func (wg *WorkGroup) Wait() error             { return wg.g.Wait() }
func (wg *WorkGroup) Go(f func() error)       { wg.g.Go(f) }
