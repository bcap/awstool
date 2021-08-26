package loader

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2Types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

func FetchAllV1ELBs(ctx context.Context, cfg aws.Config) ([]elbTypes.LoadBalancerDescription, error) {
	log.Printf("Fetching all %s ELBs (v1)", cfg.Region)
	client := elasticloadbalancing.NewFromConfig(cfg)
	describeResult, err := client.DescribeLoadBalancers(
		ctx,
		&elasticloadbalancing.DescribeLoadBalancersInput{},
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d %s elbs (v1)", len(describeResult.LoadBalancerDescriptions), cfg.Region)
	return describeResult.LoadBalancerDescriptions, nil
}

func FetchAllV2ELBs(ctx context.Context, cfg aws.Config) ([]elbv2Types.LoadBalancer, error) {
	log.Printf("Fetching all %s ELBs (v2)", cfg.Region)
	client := elasticloadbalancingv2.NewFromConfig(cfg)
	describeResult, err := client.DescribeLoadBalancers(
		ctx,
		&elasticloadbalancingv2.DescribeLoadBalancersInput{},
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d %s elbs (v2)", len(describeResult.LoadBalancers), cfg.Region)
	return describeResult.LoadBalancers, nil
}
