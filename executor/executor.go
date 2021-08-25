package executor

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Executor struct {
	wg   sync.WaitGroup
	done chan struct{}
	sem  *semaphore.Weighted
}

func NewExecutor(maxParallelism int) *Executor {
	if maxParallelism < 1 {
		maxParallelism = 1
	}
	executor := Executor{
		wg:   sync.WaitGroup{},
		done: make(chan struct{}),
		sem:  semaphore.NewWeighted(int64(maxParallelism)),
	}
	go func() {
		executor.wg.Wait()
		close(executor.done)
	}()
	return &executor
}

func (e *Executor) Launch(ctx context.Context, fn func()) <-chan struct{} {
	c := make(chan struct{})
	e.wg.Add(1)
	go func() {
		defer close(c)
		if err := e.sem.Acquire(ctx, 1); err != nil {
			return
		}
		defer e.sem.Release(1)
		fn()
		e.wg.Done()
	}()
	return c
}

func (e *Executor) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.done:
		return nil
	}
}

func (e *Executor) Done() <-chan struct{} {
	return e.done
}
