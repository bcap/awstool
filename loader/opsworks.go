package loader

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	opswTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
)

func FetchAllOpsworksStacks(ctx context.Context, cfg aws.Config) ([]opswTypes.Stack, error) {
	log.Print("Fetching all Opsworks stacks")
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeStacks(
		ctx,
		&opsworks.DescribeStacksInput{},
	)
	if err != nil {
		if isNotPresentInRegionError(err) {
			return []opswTypes.Stack{}, nil
		}
		return nil, err
	}
	log.Printf("Fetched %d Opsworks stacks", len(describeResult.Stacks))
	return describeResult.Stacks, nil
}

func FetchAllOpsworksLayers(ctx context.Context, cfg aws.Config, stackId string) ([]opswTypes.Layer, error) {
	log.Printf("Fetching all %s Opsworks layers for stack %s", cfg.Region, stackId)
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeLayers(
		ctx,
		&opsworks.DescribeLayersInput{StackId: &stackId},
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d %s Opsworks layers for stack %s", len(describeResult.Layers), cfg.Region, stackId)
	return describeResult.Layers, nil
}

func FetchAllOpsworksApps(ctx context.Context, cfg aws.Config, stackId string) ([]opswTypes.App, error) {
	log.Printf("Fetching all %s Opsworks apps for stack %s", cfg.Region, stackId)
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeApps(
		ctx,
		&opsworks.DescribeAppsInput{StackId: &stackId},
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d %s Opsworks apps for stack %s", len(describeResult.Apps), cfg.Region, stackId)
	return describeResult.Apps, nil
}

func FetchAllOpsworksInstances(ctx context.Context, cfg aws.Config, stackId string) ([]opswTypes.Instance, error) {
	log.Printf("Fetching all %s Opsworks instances for stack %s", cfg.Region, stackId)
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeInstances(
		ctx,
		&opsworks.DescribeInstancesInput{StackId: &stackId},
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d %s Opsworks instances for stack %s", len(describeResult.Instances), cfg.Region, stackId)

	return describeResult.Instances, nil
}

func isNotPresentInRegionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "dial tcp") &&
		strings.Contains(errMsg, "no such host")
}
