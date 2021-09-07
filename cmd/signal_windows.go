// +build windows

package cmd

import (
	"context"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

func CreateRunnableContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Warn("Shutting down")
		cancel()
	}()
	return ctx, cancel
}
