package loader

import (
	"strings"

	"awstool/aws/ec2"
	"awstool/aws/elasticsearch"
)

type options struct {
	includeRegions map[string]struct{}
	excludeRegions map[string]struct{}

	includeServices map[string]struct{}
	excludeServices map[string]struct{}

	ec2FetchOptions []ec2.FetchOption
	esFetchOptions  []elasticsearch.FetchOption
}

type Option func(opts *options)

func WithRegions(regions ...string) Option {
	return func(opts *options) {
		for _, region := range regions {
			opts.includeRegions[strings.ToLower(region)] = struct{}{}
		}
	}
}

func WithoutRegions(regions ...string) Option {
	return func(opts *options) {
		for _, region := range regions {
			opts.excludeRegions[strings.ToLower(region)] = struct{}{}
		}
	}
}

func WithServices(services ...string) Option {
	return func(opts *options) {
		for _, service := range services {
			opts.includeServices[strings.ToLower(service)] = struct{}{}
		}
	}
}

func WithoutServices(services ...string) Option {
	return func(opts *options) {
		for _, service := range services {
			opts.excludeServices[strings.ToLower(service)] = struct{}{}
		}
	}
}

func WithEC2FetchOptions(fetchOptions ...ec2.FetchOption) Option {
	return func(opts *options) {
		opts.ec2FetchOptions = append(opts.ec2FetchOptions, fetchOptions...)
	}
}

func WithESFetchOptions(fetchOptions ...elasticsearch.FetchOption) Option {
	return func(opts *options) {
		opts.esFetchOptions = append(opts.esFetchOptions, fetchOptions...)
	}
}

func newOptions(fns []Option) options {
	options := options{
		includeRegions:  map[string]struct{}{},
		excludeRegions:  map[string]struct{}{},
		includeServices: map[string]struct{}{},
		excludeServices: map[string]struct{}{},
		ec2FetchOptions: []ec2.FetchOption{},
	}
	for _, fn := range fns {
		fn(&options)
	}
	return options
}
