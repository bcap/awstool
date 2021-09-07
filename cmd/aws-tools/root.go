package main

import (
	"fmt"
	"os"

	awst "aws-tools/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// global var, not nice, but simplifies a lot not needing to pass data from parent
// to sub command
var awsCfg aws.Config

func RootCommand() *cobra.Command {
	cfgOptions := awst.NewAWSConfigOptions()
	var quiet bool
	var verbosity int

	cmd := cobra.Command{
		Use:           "aws-tools",
		Short:         "Set of tools to help on aws operations",
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(
		&cfgOptions.Profile, "profile", "p", "",
		"Use this AWS profile. Profiles are configured in ~/.aws/config. "+
			"If not specified then the default profile will be used",
	)

	cmd.PersistentFlags().IntVar(
		&cfgOptions.MaxRequestsInFlight, "max-requests-in-flight", cfgOptions.MaxRequestsInFlight,
		"How many requests in parallel are allowed to be executed against the AWS APIs at any"+
			"point in time",
	)

	cmd.PersistentFlags().IntVar(
		&cfgOptions.MaxRetries, "max-retries", cfgOptions.MaxRequestsInFlight,
		"Maximum amount of retries we should allow for a particular request before it fails. "+
			"See also --max-retry-time",
	)

	cmd.PersistentFlags().DurationVar(
		&cfgOptions.MaxRetryTime, "max-retry-time", cfgOptions.MaxRetryTime,
		"Maximum amount of total time we wait for a request to be retried over and over in case of failures. "+
			"See also --max-retries",
	)

	cmd.PersistentFlags().CountVarP(
		&verbosity, "verbosity", "v",
		"Controls loggging verbosity. Can be specified multiple times (eg -vv) or a count can "+
			"be passed in (--verbosity=2). Defaults to print error messages. "+
			"See also --quiet",
	)
	cmd.PersistentFlags().BoolVarP(
		&quiet, "quiet", "q", false,
		"Runs quiet, excluding even error messages. This overrides --verbosity",
	)

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := cfgOptions.Validate(); err != nil {
			return err
		}

		setupLogging(verbosity)
		if quiet {
			if err := makeItQuiet(); err != nil {
				return err
			}
		}

		log.Debugf("Starting run with the following args: %v", os.Args)

		var err error
		awsCfg, err = awst.NewAWSConfig(cmd.Context(), cfgOptions)
		if err != nil {
			return err
		}

		return nil
	}

	addSubCommand(&cmd, DumpCommand())
	addSubCommand(&cmd, ResolveCommand())

	return &cmd
}

func addSubCommand(cmd *cobra.Command, subCmd *cobra.Command) {
	if subCmd.PersistentPreRun != nil {
		innerFn := subCmd.PersistentPreRun
		subCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			if cmd.Parent().PersistentPreRun != nil {
				cmd.Parent().PersistentPreRun(cmd, args)
			}
			innerFn(cmd, args)
		}
	}
	if subCmd.PersistentPreRunE != nil {
		innerFn := subCmd.PersistentPreRunE
		subCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PersistentPreRunE != nil {
				if err := cmd.Parent().PersistentPreRunE(cmd, args); err != nil {
					return err
				}
			}
			return innerFn(cmd, args)
		}
	}
	if subCmd.PersistentPostRun != nil {
		innerFn := subCmd.PersistentPostRun
		subCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
			if cmd.Parent().PersistentPostRun != nil {
				cmd.Parent().PersistentPostRun(cmd, args)
			}
			innerFn(cmd, args)
		}
	}
	if subCmd.PersistentPostRunE != nil {
		innerFn := subCmd.PersistentPostRunE
		subCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PersistentPostRunE != nil {
				if err := cmd.Parent().PersistentPostRunE(cmd, args); err != nil {
					return err
				}
			}
			return innerFn(cmd, args)
		}
	}
	cmd.AddCommand(subCmd)
}

func setupLogging(verbosity int) {
	log.SetLevel(log.ErrorLevel)
	if verbosity == 1 {
		log.SetLevel(log.InfoLevel)
	} else if verbosity == 2 {
		log.SetLevel(log.DebugLevel)
	} else if verbosity >= 3 {
		log.SetLevel(log.TraceLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
		ForceColors:   true,
	})
	log.SetOutput(os.Stderr)
}

func makeItQuiet() error {
	// avoid logger wasting cycles generating log messages that go nowhere
	log.SetLevel(log.PanicLevel)
	f, err := os.Open("/dev/null")
	if err != nil {
		return fmt.Errorf("could not open /dev/null")
	}
	os.Stderr = f
	return nil
}