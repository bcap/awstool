package executor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
)

func TestBasic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := NewExecutor(1)
	c := make(chan struct{})

	timeoutCtx, _ := context.WithTimeout(ctx, time.Second)
	err := e.Wait(timeoutCtx)
	if err != nil {
		t.Fatal("executor Wait call should return instantly when no jobs were launched")
	}

	select {
	case <-e.Done():
	case <-time.After(time.Second):
		t.Fatal("waiting on executor Done() channel should return instantly when no jobs were launched")
	}

	e.Launch(ctx, func() {
		close(c)
	})

	select {
	case <-c:
	case <-time.After(time.Second):
		t.Fatal("simple execution should have finished instantly when waiting on job returned channel")
	}

	timeoutCtx, _ = context.WithTimeout(ctx, time.Second)
	err = e.Wait(timeoutCtx)
	if err != nil {
		t.Fatal("simple execution should have finished instantly when waiting on executor Wait call")
	}

	select {
	case <-e.Done():
	case <-time.After(time.Second):
		t.Fatal("simple execution should have finished instantly when waiting on executor Done channel")
	}
}

func TestParallelismRespected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parallelism := 10
	jobs := 500
	iterationsPerJob := 10000
	e := NewExecutor(parallelism)

	sem := semaphore.NewWeighted(int64(parallelism))
	ready := make(chan struct{})
	var executed int32
	var failed int32

	for i := 0; i < jobs; i++ {
		e.Launch(ctx, func() {
			// Work as a barrrier where all goroutines will wait on.
			// This is to minimize startup time interfering with the race
			<-ready
			for j := 0; j < iterationsPerJob; j++ {
				if !sem.TryAcquire(1) {
					atomic.CompareAndSwapInt32(&failed, 0, 1)
					return
				}
				atomic.AddInt32(&executed, 1)
				sem.Release(1)
			}
		})
	}

	close(ready) // release the barrier

	select {
	case <-e.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("executor should have finished < 10s (likely in the ms scale). Its likely deadlocked")
	}

	if failed == 1 {
		t.Fatalf("There were more concurrent goroutines running than allowed")
	}

	totalExpected := jobs * iterationsPerJob
	if int(executed) != totalExpected {
		t.Fatalf("Should have executed %d times, but actually executed %d times", totalExpected, executed)
	}
}
