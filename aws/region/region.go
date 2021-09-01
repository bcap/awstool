package region

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
)

func FetchAll(
	ctx context.Context,
	cfg aws.Config,
) ([]ec2Types.Region, error) {
	log.Debug("Fetching all regions")
	client := ec2.NewFromConfig(cfg)
	describeResult, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}
	log.Infof("Fetched %d regions", len(describeResult.Regions))
	return describeResult.Regions, nil
}
