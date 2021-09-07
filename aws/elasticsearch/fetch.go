package elasticsearch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	esTypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	log "github.com/sirupsen/logrus"
)

func ListAllDomainNames(
	ctx context.Context,
	cfg aws.Config,
) ([]string, error) {
	log.Debugf("Listing all %s Elasticsearch domain names", cfg.Region)

	client := elasticsearchservice.NewFromConfig(cfg)
	describeResult, err := client.ListDomainNames(
		ctx, &elasticsearchservice.ListDomainNamesInput{},
	)
	if err != nil {
		return nil, err
	}
	log.Infof(
		"Listed %d %s Elasticsearch domain names",
		len(describeResult.DomainNames), cfg.Region,
	)
	result := make([]string, len(describeResult.DomainNames))
	for idx, domainName := range describeResult.DomainNames {
		result[idx] = *domainName.DomainName
	}
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
	log.Debugf("Fetching %s Elasticsearch domain config for domain %s", cfg.Region, domain)

	client := elasticsearchservice.NewFromConfig(cfg)
	describeResult, err := client.DescribeElasticsearchDomainConfig(
		ctx, &elasticsearchservice.DescribeElasticsearchDomainConfigInput{
			DomainName: &domain,
		},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %s Elasticsearch domain config for domain %s", cfg.Region, domain)
	return describeResult.DomainConfig, nil
}
