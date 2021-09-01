package loader

import (
	"strings"
)

type options struct {
	includeRegions map[string]struct{}
	excludeRegions map[string]struct{}

	includeServices map[string]struct{}
	excludeServices map[string]struct{}
}

type Option func(opts *options)

func WithRegions(regions []string) Option {
	return func(opts *options) {
		for _, region := range regions {
			opts.includeRegions[strings.ToLower(region)] = struct{}{}
		}
	}
}

func WithoutRegions(regions []string) Option {
	return func(opts *options) {
		for _, region := range regions {
			opts.excludeRegions[strings.ToLower(region)] = struct{}{}
		}
	}
}

func WithServices(services []string) Option {
	return func(opts *options) {
		for _, service := range services {
			opts.includeServices[strings.ToLower(service)] = struct{}{}
		}
	}
}

func WithoutServices(services []string) Option {
	return func(opts *options) {
		for _, service := range services {
			opts.excludeServices[strings.ToLower(service)] = struct{}{}
		}
	}
}

func newOptions(fns []Option) options {
	options := options{
		includeRegions:  map[string]struct{}{},
		excludeRegions:  map[string]struct{}{},
		includeServices: map[string]struct{}{},
		excludeServices: map[string]struct{}{},
	}
	for _, fn := range fns {
		fn(&options)
	}
	return options
}
