package aws

import (
	"context"
	"fmt"
	"time"

	"awstool/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	smithylogging "github.com/aws/smithy-go/logging"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

const defaultMaxRequestsInFlight = 50
const defaultMaxRetries = 9
const defaultMaxRetryTime = 10 * time.Second

type AWSConfigOptions struct {
	Profile             string
	MaxRequestsInFlight int
	MaxRetries          int
	MaxRetryTime        time.Duration
}

func NewAWSConfigOptions() AWSConfigOptions {
	return AWSConfigOptions{
		MaxRequestsInFlight: defaultMaxRequestsInFlight,
		MaxRetries:          defaultMaxRetries,
		MaxRetryTime:        defaultMaxRetryTime,
	}
}

func (o *AWSConfigOptions) Validate() error {
	if o.MaxRequestsInFlight < 1 {
		return fmt.Errorf("maxRequestsInFlight must be higher than zero (passed: %d)", o.MaxRequestsInFlight)
	}
	if o.MaxRetries < 0 {
		return fmt.Errorf("maxRetries must be a positive integer (passed: %d)", o.MaxRetries)
	}
	if o.MaxRetryTime < 0 {
		return fmt.Errorf("maxRetryTime must be a positive time (passed: %v)", o.MaxRetryTime)
	}
	return nil
}

func NewAWSConfig(ctx context.Context, options AWSConfigOptions) (aws.Config, error) {
	if err := options.Validate(); err != nil {
		return aws.Config{}, err
	}

	log.Debug(spew.Sprintf("Creating config with the following options: %+v", options))

	httpClient := http.NewParallelLimitedHTTPClient(
		awshttp.NewBuildableClient(),
		options.MaxRequestsInFlight,
	)

	createRetryer := func() aws.Retryer {
		return retry.NewStandard(
			func(opts *retry.StandardOptions) {
				opts.MaxAttempts = options.MaxRetries
				opts.MaxBackoff = options.MaxRetryTime
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

	if options.Profile != "" {
		cfgOptions = append(cfgOptions, config.WithSharedConfigProfile(options.Profile))
	}

	return config.LoadDefaultConfig(ctx, cfgOptions...)
}
