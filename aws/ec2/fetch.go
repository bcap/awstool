package ec2

import (
	"context"
	"strings"

	"awstool/common"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
)

type FetchOption func(opt *ec2.DescribeInstancesInput)

func WithInstanceIds(instanceIds ...string) FetchOption {
	return func(opt *ec2.DescribeInstancesInput) {
		opt.InstanceIds = append(opt.InstanceIds, instanceIds...)
	}
}

func WithTag(key string, value string) FetchOption {
	return WithFilter("tag:"+key, value)
}

func WithFilter(name string, values ...string) FetchOption {
	return func(opt *ec2.DescribeInstancesInput) {
		opt.Filters = append(opt.Filters, ec2Types.Filter{
			Name:   &name,
			Values: values,
		})
	}
}

func WithFilters(filters ...ec2Types.Filter) FetchOption {
	return func(opt *ec2.DescribeInstancesInput) {
		opt.Filters = append(opt.Filters, filters...)
	}
}

func newDescribeInput(opts ...FetchOption) *ec2.DescribeInstancesInput {
	result := ec2.DescribeInstancesInput{
		InstanceIds: []string{},
		Filters:     []ec2Types.Filter{},
	}
	for _, fn := range opts {
		fn(&result)
	}
	return &result
}

func FetchAllInstances(
	ctx context.Context,
	cfg aws.Config,
	options ...FetchOption,
) ([]ec2Types.Reservation, error) {
	log.Debugf("Fetching all EC2 reservations and instances in %s", cfg.Region)

	reservations := []ec2Types.Reservation{}
	numInstances := 0
	numGroups := 0

	client := ec2.NewFromConfig(cfg)
	describeInput := newDescribeInput(options...)

	load := func(nextToken *string) (*string, error) {
		describeInput.NextToken = nextToken
		describeResult, err := client.DescribeInstances(ctx, describeInput)
		if err != nil {
			if len(describeInput.InstanceIds) > 0 && isInstanceNotFoundError(err) {
				return nil, nil
			}
			return nil, err
		}
		for _, reservation := range describeResult.Reservations {
			numInstances += len(reservation.Instances)
			numGroups += len(reservation.Groups)
		}
		reservations = append(reservations, describeResult.Reservations...)
		return describeResult.NextToken, nil
	}

	err := common.FetchAll("instances", load)
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

	err := common.FetchAll("ebs volumes", load)
	if err != nil {
		return volumes, err
	}

	log.Infof("Fetched %d %s volumes", len(volumes), cfg.Region)

	return volumes, nil
}

func isInstanceNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "error InvalidInstanceID.NotFound")
}
