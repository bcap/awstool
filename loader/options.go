package loader

import (
	"time"
)

// Options

const (
	DefaultMaxRetries          = 5
	DefaultSleepBetweenRetries = 200 * time.Millisecond
	DefaultBackoffFactor       = 2.0
)

type loaderOptions struct {
	maxRetries          int
	sleepBetweenRetries time.Duration
	backoffFactor       float32
}

func newLoaderOptions(options ...LoaderOptionFn) loaderOptions {
	opts := loaderOptions{
		maxRetries:          DefaultMaxRetries,
		sleepBetweenRetries: DefaultSleepBetweenRetries,
		backoffFactor:       DefaultBackoffFactor,
	}

	// apply option functions
	for _, optFn := range options {
		optFn(&opts)
	}

	// sanitize the input applying sensible values instead of
	// strictly returning errors
	if opts.maxRetries < 0 {
		opts.maxRetries = 0
	}
	if opts.sleepBetweenRetries < 0 {
		opts.sleepBetweenRetries = 0
	}
	if opts.backoffFactor < 1.0 {
		opts.backoffFactor = 1.0
	}

	return opts
}

type LoaderOptionFn = func(*loaderOptions)

func WithMaxRetries(maxRetries int) LoaderOptionFn {
	return func(opts *loaderOptions) {
		opts.maxRetries = maxRetries
	}
}

func WithSleepBetweenRetries(sleep time.Duration) LoaderOptionFn {
	return func(opts *loaderOptions) {
		opts.sleepBetweenRetries = sleep
	}
}

func WithBackoffFactor(factor float32) LoaderOptionFn {
	return func(opts *loaderOptions) {
		opts.backoffFactor = factor
	}
}
