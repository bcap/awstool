package loader

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	ebtTypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	log "github.com/sirupsen/logrus"
)

func FetchAllElasticBeanstalkApplications(
	ctx context.Context,
	cfg aws.Config,
) ([]ebtTypes.ApplicationDescription, error) {
	log.Debugf("Fetching all %s Elastic Beanstalk applications", cfg.Region)

	client := elasticbeanstalk.NewFromConfig(cfg)
	describeResult, err := client.DescribeApplications(
		ctx, &elasticbeanstalk.DescribeApplicationsInput{},
	)
	if err != nil {
		return nil, err
	}
	log.Infof(
		"Fetched %d %s Elastic Beanstalk applications",
		len(describeResult.Applications), cfg.Region,
	)
	return describeResult.Applications, nil
}

func FetchAllElasticBeanstalkEnvironments(
	ctx context.Context,
	cfg aws.Config,
) ([]ebtTypes.EnvironmentDescription, error) {
	log.Debugf("Fetching all %s Elastic Beanstalk environments", cfg.Region)

	client := elasticbeanstalk.NewFromConfig(cfg)
	describeResult, err := client.DescribeEnvironments(
		ctx, &elasticbeanstalk.DescribeEnvironmentsInput{},
	)
	if err != nil {
		return nil, err
	}
	log.Infof(
		"Fetched %d %s Elastic Beanstalk environments",
		len(describeResult.Environments), cfg.Region,
	)
	return describeResult.Environments, nil
}
