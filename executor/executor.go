package executor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

type Executor struct {
	wg             sync.WaitGroup
	sem            *semaphore.Weighted
	launchedAtomic int32
}

// NewExecutor creates a new executor limited to run at most maxParallelism operations
// in parallel. For ubounded operations in parallel, set maxParallelism to zero or below
func NewExecutor(maxParallelism int) *Executor {
	var sem *semaphore.Weighted
	if maxParallelism > 0 {
		sem = semaphore.NewWeighted(int64(maxParallelism))
	}
	executor := Executor{
		wg:  sync.WaitGroup{},
		sem: sem,
	}
	log.Tracef("Created executor %p with max parallelism of %d", &executor, maxParallelism)
	return &executor
}

func (e *Executor) Launch(ctx context.Context, fn func()) <-chan struct{} {
	c := make(chan struct{})

	e.wg.Add(1)
	fnId := atomic.AddInt32(&e.launchedAtomic, 1) - 1

	queuedAt := time.Now()
	log.Tracef("Executor %p queued function #%d: %#+v", e, fnId, fn)

	go func() {
		defer close(c)
		defer e.wg.Done()

		if e.sem != nil {
			// sem.Acquire will only return an error if the context gets cancelled
			// For that reason we do not bother checking the error
			if err := e.sem.Acquire(ctx, 1); err != nil {
				return
			}
			defer e.sem.Release(1)
		}

		runningAt := time.Now()
		log.Tracef("Executor %p running function #%d: %#+v (waited %v in queue)", e, fnId, fn, runningAt.Sub(queuedAt))

		fn()

		finishedAt := time.Now()
		log.Tracef("Executor %p finished running function #%d: %#+v (executed in %v)", e, fnId, fn, finishedAt.Sub(runningAt))
	}()

	return c
}

func (e *Executor) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()
	return done
}

func (e *Executor) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.Done():
		return nil
	}
}
