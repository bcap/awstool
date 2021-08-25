package ec2

import (
	"context"
	"log"

	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func FetchAllInstances(
	ctx context.Context,
	cfg aws.Config,
	loaderOpts ...loader.LoaderOptionFn,
) ([]ec2Types.Reservation, error) {
	log.Print("Fetching all EC2 instances")

	reservations := []ec2Types.Reservation{}
	batches := 0
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
		batches++
		return describeResult.NextToken, nil
	}

	err := loader.FetchAll("instances", load, loaderOpts...)
	if err != nil {
		return reservations, err
	}

	log.Printf(
		"Fetched %d reservations with a total of %d groups and %d instances in %d batches",
		len(reservations), numGroups, numInstances, batches,
	)
	return reservations, nil
}

func FetchAllEBSVolumes(
	ctx context.Context,
	cfg aws.Config,
	loaderOpts ...loader.LoaderOptionFn,
) ([]ec2Types.Volume, error) {
	log.Print("Fetching all EC2 EBS volumes")

	volumes := []ec2Types.Volume{}
	batches := 0

	client := ec2.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, describeResult.Volumes...)
		batches++
		return describeResult.NextToken, nil
	}

	err := loader.FetchAll("ebs volumes", load, loaderOpts...)
	if err != nil {
		return volumes, err
	}

	log.Printf("Fetched %d volumes in %d batches", len(volumes), batches)

	return volumes, nil
}
