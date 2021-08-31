package cmd

import (
	"context"
	"os"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

func LogMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	toMb := func(x uint64) int64 {
		return int64(x / uint64(1024) / uint64(1024))
	}
	log.Debugf("Mem stats: Pid=%d, Alloc=%dMiB, TotalAlloc=%dMiB, GCRuns=%d",
		os.Getpid(), toMb(m.Alloc), toMb(m.TotalAlloc), m.NumGC,
	)
}

func KeepLoggingMemoryUsage(ctx context.Context, period time.Duration) {
	go func() {
		LogMemoryUsage()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(period):
				LogMemoryUsage()
			}
		}
	}()
}
