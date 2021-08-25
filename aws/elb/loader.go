package elb

import (
	"context"
	"log"

	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

func FetchAllV1ELBs(
	ctx context.Context,
	cfg aws.Config,
	loaderOpts ...loader.LoaderOptionFn,
) ([]elbTypes.LoadBalancerDescription, error) {
	log.Print("Fetching all ELBs (v1)")

	result := []elbTypes.LoadBalancerDescription{}
	batches := 0

	client := elasticloadbalancing.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeLoadBalancers(
			ctx,
			&elasticloadbalancing.DescribeLoadBalancersInput{},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.LoadBalancerDescriptions...)
		batches++
		return nil, nil
	}

	err := loader.FetchAll("elbs (v1)", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d elbs (v1) in %d batches", len(result), batches)

	return result, nil
}

func FetchAllV2ELBs(
	ctx context.Context,
	cfg aws.Config,
	loaderOpts ...loader.LoaderOptionFn,
) ([]elbv2Types.LoadBalancer, error) {
	log.Print("Fetching all ELBs (v2)")

	result := []elbv2Types.LoadBalancer{}
	batches := 0

	client := elasticloadbalancingv2.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeLoadBalancers(
			ctx,
			&elasticloadbalancingv2.DescribeLoadBalancersInput{},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.LoadBalancers...)
		batches++
		return nil, nil
	}

	err := loader.FetchAll("elbs (v2)", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d elbs (v2) in %d batches", len(result), batches)

	return result, nil
}
