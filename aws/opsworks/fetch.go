package opsworks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	opswTypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	log "github.com/sirupsen/logrus"
)

func FetchAllStacks(ctx context.Context, cfg aws.Config) ([]opswTypes.Stack, error) {
	log.Debugf("Fetching all %s Opsworks stacks", cfg.Region)
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
	log.Infof("Fetched %d %s Opsworks stacks", len(describeResult.Stacks), cfg.Region)
	return describeResult.Stacks, nil
}

func FetchAllLayers(ctx context.Context, cfg aws.Config, stackId string) ([]opswTypes.Layer, error) {
	log.Debugf("Fetching all %s Opsworks layers for stack %s", cfg.Region, stackId)
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeLayers(
		ctx,
		&opsworks.DescribeLayersInput{StackId: &stackId},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %d %s Opsworks layers for stack %s", len(describeResult.Layers), cfg.Region, stackId)
	return describeResult.Layers, nil
}

func FetchAllApps(ctx context.Context, cfg aws.Config, stackId string) ([]opswTypes.App, error) {
	log.Debugf("Fetching all %s Opsworks apps for stack %s", cfg.Region, stackId)
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeApps(
		ctx,
		&opsworks.DescribeAppsInput{StackId: &stackId},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %d %s Opsworks apps for stack %s", len(describeResult.Apps), cfg.Region, stackId)
	return describeResult.Apps, nil
}

func FetchAllInstances(ctx context.Context, cfg aws.Config, stackId string) ([]opswTypes.Instance, error) {
	log.Debugf("Fetching all %s Opsworks instances for stack %s", cfg.Region, stackId)
	client := opsworks.NewFromConfig(cfg)
	describeResult, err := client.DescribeInstances(
		ctx,
		&opsworks.DescribeInstancesInput{StackId: &stackId},
	)
	if err != nil {
		return nil, err
	}
	log.Debugf("Fetched %d %s Opsworks instances for stack %s", len(describeResult.Instances), cfg.Region, stackId)

	return describeResult.Instances, nil
}

func isNotPresentInRegionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "dial tcp") &&
		strings.Contains(errMsg, "no such host")
}
