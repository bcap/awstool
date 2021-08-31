package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"aws-tools/cmd"
	dump "aws-tools/dump"
	dumphttp "aws-tools/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	smithylogging "github.com/aws/smithy-go/logging"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := cmd.CreateRunnableContext()
	defer cancel()

	cmd.KeepLoggingMemoryUsage(ctx, 15*time.Second)

	cmd := createCommand()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

type Options struct {
	awsProfile          string
	maxRequestsInFlight int
	maxRetries          int
	maxRetryTime        time.Duration
	dumpOptions         []dump.Option
}

func (o *Options) validate() error {
	if o.maxRequestsInFlight < 1 {
		return fmt.Errorf("maxRequestsInFlight must be higher than zero (passed: %d)", o.maxRequestsInFlight)
	}
	if o.maxRetries < 0 {
		return fmt.Errorf("maxRetries must be a positive integer (passed: %d)", o.maxRetries)
	}
	if o.maxRetryTime < 0 {
		return fmt.Errorf("maxRetryTime must be a positive time (passed: %v)", o.maxRetryTime)
	}
	return nil
}

const defaultMaxRequestsInFlight = 50
const defaultMaxRetries = 9
const defaultMaxRetryTime = 10 * time.Second

func createCommand() *cobra.Command {
	options := Options{}
	var quiet bool
	var verbosity int
	var regions []string
	var excludeRegions []string
	var services []string
	var excludeServices []string

	cmd := cobra.Command{
		Use:           "dump",
		Short:         "Generates a single json dumping the results of many different description APIs from AWS",
		SilenceErrors: true,
	}

	cmd.Flags().StringVarP(
		&options.awsProfile, "profile", "p", "",
		"Use this AWS profile. Profiles are configured in ~/.aws/config. "+
			"If not specified then the default profile will be used",
	)

	cmd.Flags().IntVar(
		&options.maxRequestsInFlight, "max-requests-in-flight", defaultMaxRequestsInFlight,
		"How many requests in parallel are allowed to be executed against the AWS APIs at any"+
			"point in time",
	)

	cmd.Flags().IntVar(
		&options.maxRetries, "max-retries", defaultMaxRetries,
		"Maximum amount of retries we should allow for a particular request before it fails. "+
			"See also --max-retry-time",
	)

	cmd.Flags().DurationVar(
		&options.maxRetryTime, "max-retry-time", defaultMaxRetryTime,
		"Maximum amount of total time we wait for a request to be retried over and over in case of failures. "+
			"See also --max-retries",
	)

	cmd.Flags().StringSliceVarP(
		&regions, "regions", "r", []string{},
		"Dump data for only those regions. If not specified, all regions will be dumped. "+
			"See also --exclude-regions",
	)
	cmd.Flags().StringSliceVarP(
		&excludeRegions, "exclude-regions", "R", []string{},
		"Do not dump data for those regions. This takes precedence over --regions",
	)

	cmd.Flags().StringSliceVarP(
		&services, "services", "s", []string{},
		"Dump only those services. If not specified, all implemented services will be dumped. "+
			"See also --exclude-services",
	)
	cmd.Flags().StringSliceVarP(
		&excludeServices, "exclude-services", "S", []string{},
		"Do not dump data for those services. This takes precedence over --services",
	)

	cmd.Flags().CountVarP(
		&verbosity, "verbosity", "v",
		"Controls loggging verbosity. Can be specified multiple times (eg -vv) or a count can "+
			"be passed in (--verbosity=2). Defaults to print error messages. "+
			"See also --quiet",
	)
	cmd.Flags().BoolVarP(
		&quiet, "quiet", "q", false,
		"Runs quiet, excluding even error messages. This overrides --verbosity",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		setupLogging(verbosity)
		if quiet {
			if err := makeItQuiet(&options); err != nil {
				return err
			}
		}

		options.dumpOptions = []dump.Option{
			dump.WithRegions(regions),
			dump.WithoutRegions(excludeRegions),
			dump.WithServices(services),
			dump.WithoutServices(excludeServices),
		}
		if err := options.validate(); err != nil {
			return err
		}

		// we silence usage here instead of setting in the command struct declaration because it is
		// only at this point forward that we want to not display the usage when an error occurs,
		// as it will be an execution error, not a parsing/usage error
		// see more at https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		return run(cmd.Context(), options)
	}

	return &cmd
}

func run(ctx context.Context, options Options) error {
	cfg, err := createAWSConfig(ctx, options)
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		return err
	}

	result, err := dump.DumpAWS(ctx, cfg, options.dumpOptions...)
	if err != nil {
		log.Errorf("Execution failed: %v", err)
		return err
	}

	log.Info("Data fully loaded, encoding to json")
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		log.Errorf("Error while encoding result to json: %v", err)
		return err
	}

	fmt.Println(string(jsonBytes))

	log.Debug("Done")

	return nil
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
	})
	log.SetOutput(os.Stderr)
}

func createAWSConfig(ctx context.Context, options Options) (aws.Config, error) {
	httpClient := dumphttp.NewParallelLimitedHTTPClient(
		awshttp.NewBuildableClient(),
		options.maxRequestsInFlight,
	)

	createRetryer := func() aws.Retryer {
		return retry.NewStandard(
			func(opts *retry.StandardOptions) {
				opts.MaxAttempts = options.maxRetries
				opts.MaxBackoff = options.maxRetryTime
			},
		)
	}

	logger := log.New()
	wrappedLogger := smithylogging.LoggerFunc(
		func(classification smithylogging.Classification, format string, v ...interface{}) {
			lv, err := log.ParseLevel(string(classification))
			if err != nil {
				lv = log.InfoLevel
			}
			logger.Logf(lv, format, v...)
		},
	)

	cfgOptions := []func(*config.LoadOptions) error{
		config.WithHTTPClient(httpClient),
		config.WithRetryer(createRetryer),
		config.WithLogger(wrappedLogger),
	}

	if options.awsProfile != "" {
		cfgOptions = append(cfgOptions, config.WithSharedConfigProfile(options.awsProfile))
	}

	return config.LoadDefaultConfig(ctx, cfgOptions...)
}

func makeItQuiet(options *Options) error {
	// avoid logger wasting cycles generating log messages that go nowhere
	log.SetLevel(log.PanicLevel)
	f, err := os.Open("/dev/null")
	if err != nil {
		return fmt.Errorf("could not open /dev/null")
	}
	os.Stderr = f
	return nil
}
