package main

import (
	"os"
	"time"

	"awstool/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := cmd.CreateRunnableContext()
	defer cancel()

	cmd.KeepLoggingMemoryUsage(ctx, 15*time.Second, log.TraceLevel)

	if err := RootCommand().ExecuteContext(ctx); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Debug("Command finished successfully")
}
