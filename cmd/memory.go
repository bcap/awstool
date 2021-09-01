package cmd

import (
	"context"
	"os"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

func LogMemoryUsage(level log.Level) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	toMb := func(x uint64) int64 {
		return int64(x / uint64(1024) / uint64(1024))
	}
	log.StandardLogger().Logf(
		level,
		"Mem stats: Pid=%d, Alloc=%dMiB, TotalAlloc=%dMiB, GCRuns=%d",
		os.Getpid(), toMb(m.Alloc), toMb(m.TotalAlloc), m.NumGC,
	)
}

func KeepLoggingMemoryUsage(ctx context.Context, period time.Duration, level log.Level) {
	go func() {
		LogMemoryUsage(level)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(period):
				LogMemoryUsage(level)
			}
		}
	}()
}
