package main

import (
	"fmt"
	"os"

	awst "awstool/aws"
	awstcmd "awstool/cmd"
	"awstool/cmd/awstool/dump"
	"awstool/cmd/awstool/ec2"
	"awstool/cmd/awstool/es"
	"awstool/cmd/awstool/s3"

	"github.com/aws/aws-sdk-go-v2/aws"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func RootCommand() *cobra.Command {
	cfgOptions := awst.NewAWSConfigOptions()
	var quiet bool
	var verbosity int

	cmd := cobra.Command{
		Use:           "awstool",
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

	var awsCfgP *aws.Config

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

		awsCfg, err := awst.NewAWSConfig(cmd.Context(), cfgOptions)
		awsCfgP = &awsCfg
		if err != nil {
			return err
		}
		return nil
	}

	awstcmd.AddSubCommand(&cmd, dump.Command(&awsCfgP))
	awstcmd.AddSubCommand(&cmd, ec2.Command(&awsCfgP))
	awstcmd.AddSubCommand(&cmd, es.Command(&awsCfgP))
	awstcmd.AddSubCommand(&cmd, s3.Command(&awsCfgP))

	return &cmd
}

func setupLogging(verbosity int) {
	log.SetLevel(log.WarnLevel)
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
