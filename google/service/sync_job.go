package service

import (
	"context"
	"sync"

	"github.com/gaucho-racing/sentinel/google/pkg/logger"
)

// syncJob serializes background work with "latest-wins" semantics. Calling
// Start cancels any in-flight run and queues a new one that begins as soon as
// the cancelled run has exited.
//
// The pattern is safe specifically because the reconcile op is idempotent:
// every run reads the live state, computes a diff, and applies it — so a run
// that gets cancelled mid-way is harmless, and the next run catches up from
// whatever state the world ended up in. fn is expected to check ctx.Err() at
// convenient points (between bindings, between member writes); there's no
// attempt to abort an in-flight HTTP request.
type syncJob struct {
	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

// Start cancels any in-flight run and spawns a new one with fn. Returns
// immediately once the new goroutine is queued (does not wait for it to
// finish). Successive rapid calls each cancel the previous; only the most
// recent fn is guaranteed to run to completion.
func (sj *syncJob) Start(fn func(ctx context.Context)) {
	sj.mu.Lock()

	// Cancel the previous run (if any) and remember its done so the new
	// goroutine can wait for the old one to fully exit before starting —
	// that's what gives us true serialization (no overlapping writes).
	if sj.cancel != nil {
		sj.cancel()
	}
	prevDone := sj.done

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	sj.cancel = cancel
	sj.done = done

	sj.mu.Unlock()

	go func() {
		// Close `done` last so chained waiters know we've fully exited.
		defer close(done)

		// Recover so a panic in fn doesn't deadlock future Starts (the
		// goroutine would die before close(done), waiters would block
		// forever, and sj.cancel would stay non-nil).
		defer func() {
			if r := recover(); r != nil {
				logger.SugarLogger.Errorf("sync job panic: %v", r)
			}
		}()

		// Wait for previous run to exit fully before starting our work.
		if prevDone != nil {
			<-prevDone
		}

		// fn must respect ctx — if we were already cancelled by a newer
		// Start before getting here, fn's first ctx.Err() check exits it.
		fn(ctx)

		// Clear our pointers ONLY if we're still the latest. If a newer
		// Start replaced us, sj.done points at its `done`, not ours, and
		// we mustn't clobber it.
		sj.mu.Lock()
		if sj.done == done {
			sj.cancel = nil
			sj.done = nil
		}
		sj.mu.Unlock()
	}()
}
