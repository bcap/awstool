package loader

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func FetchAllRegions(
	ctx context.Context,
	cfg aws.Config,
) ([]ec2Types.Region, error) {
	log.Print("Fetching all regions")
	client := ec2.NewFromConfig(cfg)
	describeResult, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d regions", len(describeResult.Regions))
	return describeResult.Regions, nil
}
