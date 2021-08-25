package opsworks

import (
	"context"
	"log"

	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	opswTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
)

func FetchAllOpsworksStacks(
	ctx context.Context,
	cfg aws.Config,
	loaderOpts ...loader.LoaderOptionFn,
) ([]opswTypes.Stack, error) {
	log.Print("Fetching all Opsworks stacks")

	result := []opswTypes.Stack{}

	client := opsworks.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeStacks(
			ctx,
			&opsworks.DescribeStacksInput{},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.Stacks...)
		return nil, nil
	}

	err := loader.FetchAll("Opsworks stacks", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d Opsworks stacks", len(result))

	return result, nil
}

func FetchAllOpsworksLayers(
	ctx context.Context,
	cfg aws.Config,
	stackId string,
	loaderOpts ...loader.LoaderOptionFn,
) ([]opswTypes.Layer, error) {
	log.Printf("Fetching all Opsworks layers for stack %s", stackId)

	result := []opswTypes.Layer{}

	client := opsworks.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeLayers(
			ctx,
			&opsworks.DescribeLayersInput{StackId: &stackId},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.Layers...)
		return nil, nil
	}

	err := loader.FetchAll("Opsworks layers", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d Opsworks layers for stack %s", len(result), stackId)

	return result, nil
}

func FetchAllOpsworksApps(
	ctx context.Context,
	cfg aws.Config,
	stackId string,
	loaderOpts ...loader.LoaderOptionFn,
) ([]opswTypes.App, error) {
	log.Printf("Fetching all Opsworks apps for stack %s", stackId)

	result := []opswTypes.App{}

	client := opsworks.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {

		describeResult, err := client.DescribeApps(
			ctx,
			&opsworks.DescribeAppsInput{StackId: &stackId},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.Apps...)
		return nil, nil
	}

	err := loader.FetchAll("Opsworks apps", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d Opsworks apps for stack %s", len(result), stackId)

	return result, nil
}

func FetchAllOpsworksInstances(
	ctx context.Context,
	cfg aws.Config,
	stackId string,
	loaderOpts ...loader.LoaderOptionFn,
) ([]opswTypes.Instance, error) {
	log.Printf("Fetching all Opsworks instances for stack %s", stackId)

	result := []opswTypes.Instance{}

	client := opsworks.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.DescribeInstances(
			ctx,
			&opsworks.DescribeInstancesInput{StackId: &stackId},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.Instances...)
		return nil, nil
	}

	err := loader.FetchAll("Opsworks instances", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d Opsworks instances for stack %s", len(result), stackId)

	return result, nil
}
