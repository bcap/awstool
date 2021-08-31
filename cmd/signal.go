package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func CreateRunnableContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		notified := false
		for range c {
			if !notified {
				log.Warn("Shutting down. Press ^C a second time to forcefully shutdown")
				notified = true
				cancel()
			} else {
				log.Warn("Forcefully shutting down (SIGKILL)")
				err := syscall.Kill(os.Getpid(), syscall.SIGKILL)
				if err != nil {
					log.Panicf("Failed to kill: %v", err)
				}
			}
		}
	}()
	return ctx, cancel
}
