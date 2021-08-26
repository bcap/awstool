package executor

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Executor struct {
	wg  sync.WaitGroup
	sem *semaphore.Weighted
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
	return &executor
}

func (e *Executor) Launch(ctx context.Context, fn func()) <-chan struct{} {
	c := make(chan struct{})
	e.wg.Add(1)
	go func() {
		defer close(c)
		if e.sem != nil {
			// sem.Acquire will only return an error if the context gets cancelled
			// For that reason we do not bother checking the error
			if err := e.sem.Acquire(ctx, 1); err != nil {
				return
			}
			defer e.sem.Release(1)
		}
		fn()
		e.wg.Done()
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
