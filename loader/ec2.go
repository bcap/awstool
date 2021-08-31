package loader

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
)

func FetchAllInstances(
	ctx context.Context,
	cfg aws.Config,
) ([]ec2Types.Reservation, error) {
	log.Debugf("Fetching all EC2 reservations and instances in %s", cfg.Region)

	reservations := []ec2Types.Reservation{}
	numInstances := 0
	numGroups := 0

	client := ec2.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		for _, reservation := range describeResult.Reservations {
			numInstances += len(reservation.Instances)
			numGroups += len(reservation.Groups)
		}
		reservations = append(reservations, describeResult.Reservations...)
		return describeResult.NextToken, nil
	}

	err := FetchAll("instances", load)
	if err != nil {
		return reservations, err
	}

	log.Infof(
		"Fetched %d %s reservations with a total of %d groups and %d instances",
		len(reservations), cfg.Region, numGroups, numInstances,
	)
	return reservations, nil
}

func FetchAllEBSVolumes(
	ctx context.Context,
	cfg aws.Config,
) ([]ec2Types.Volume, error) {
	log.Debugf("Fetching all %s EBS volumes", cfg.Region)

	volumes := []ec2Types.Volume{}

	client := ec2.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, describeResult.Volumes...)
		return describeResult.NextToken, nil
	}

	err := FetchAll("ebs volumes", load)
	if err != nil {
		return volumes, err
	}

	log.Infof("Fetched %d %s volumes", len(volumes), cfg.Region)

	return volumes, nil
}
