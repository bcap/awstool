package elasticsearch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	esTypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	log "github.com/sirupsen/logrus"
)

type fetchOptions struct {
	domains map[string]struct{}
}

func newFetchOptions(opts ...FetchOption) fetchOptions {
	options := fetchOptions{
		domains: map[string]struct{}{},
	}
	for _, fn := range opts {
		fn(&options)
	}
	return options
}

type FetchOption func(opt *fetchOptions)

func WithDomains(domains ...string) FetchOption {
	return func(opt *fetchOptions) {
		for _, domain := range domains {
			opt.domains[domain] = struct{}{}
		}
	}
}

func ListAllDomainNames(
	ctx context.Context,
	cfg aws.Config,
	fetchoptions ...FetchOption,
) ([]string, error) {
	opts := newFetchOptions(fetchoptions...)
	log.Debugf("Listing all %s Elasticsearch domain names", cfg.Region)

	client := elasticsearchservice.NewFromConfig(cfg)
	describeResult, err := client.ListDomainNames(
		ctx, &elasticsearchservice.ListDomainNamesInput{},
	)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(describeResult.DomainNames))
	for idx, domainName := range describeResult.DomainNames {
		result[idx] = *domainName.DomainName
	}

	// filter out domains if specified
	if len(opts.domains) > 0 {
		filteredResult := []string{}
		for _, domain := range result {
			if _, ok := opts.domains[domain]; ok {
				filteredResult = append(filteredResult, domain)
			}
		}
		result = filteredResult
	}

	log.Infof(
		"Listed %d %s Elasticsearch domain names",
		len(result), cfg.Region,
	)

	return result, nil
}

func FetchDomainStatus(
	ctx context.Context,
	cfg aws.Config,
	domain string,
) (*esTypes.ElasticsearchDomainStatus, error) {
	log.Debugf("Fetching %s Elasticsearch domain %s", cfg.Region, domain)

	client := elasticsearchservice.NewFromConfig(cfg)
	describeResult, err := client.DescribeElasticsearchDomain(
		ctx, &elasticsearchservice.DescribeElasticsearchDomainInput{
			DomainName: &domain,
		},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %s Elasticsearch domain %s", cfg.Region, domain)
	return describeResult.DomainStatus, nil
}

func FetchDomainConfig(
	ctx context.Context,
	cfg aws.Config,
	domain string,
) (*esTypes.ElasticsearchDomainConfig, error) {
	log.Debugf("Fetching %s Elasticsearch domain config for %s", cfg.Region, domain)

	client := elasticsearchservice.NewFromConfig(cfg)
	describeResult, err := client.DescribeElasticsearchDomainConfig(
		ctx, &elasticsearchservice.DescribeElasticsearchDomainConfigInput{
			DomainName: &domain,
		},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %s Elasticsearch domain config for %s", cfg.Region, domain)
	return describeResult.DomainConfig, nil
}

func FetchDomainTags(
	ctx context.Context,
	cfg aws.Config,
	domainARN string,
) ([]esTypes.Tag, error) {
	log.Debugf("Fetching %s Elasticsearch domain tags for %s", cfg.Region, domainARN)

	client := elasticsearchservice.NewFromConfig(cfg)
	result, err := client.ListTags(
		ctx, &elasticsearchservice.ListTagsInput{
			ARN: &domainARN,
		},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %s Elasticsearch domain tags for %s", cfg.Region, domainARN)
	return result.TagList, nil
}
