package main

import (
	"os"
	"time"

	"awstool/cmd"

	log "github.com/sirupsen/logrus"
)

type exitCodeErr interface {
	ExitCode() int
}

func main() {
	ctx, cancel := cmd.CreateRunnableContext()
	defer cancel()

	cmd.KeepLoggingMemoryUsage(ctx, 15*time.Second, log.TraceLevel)

	if err := RootCommand().ExecuteContext(ctx); err != nil {
		log.Error(err)
		exitCode := 1
		if errWithCode, ok := err.(exitCodeErr); ok {
			exitCode = errWithCode.ExitCode()
		}
		os.Exit(exitCode)
	}

	log.Debug("Command finished successfully")
}
