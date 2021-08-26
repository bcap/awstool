package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"aws-tools/common"
	"aws-tools/executor"
	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func DumpAWS(ctx context.Context, cfg aws.Config) (AWS, error) {
	result := NewAWS()
	errors := []error{}
	var resultLock sync.Mutex

	regions, err := loader.FetchAllRegions(ctx, cfg)
	if err != nil {
		return result, err
	}

	executor := executor.NewExecutor(0)
	for _, region := range regions {
		regionName := *region.RegionName
		executor.Launch(ctx, func() {
			cfgCopy := cfg
			cfgCopy.Region = regionName
			regionDump, err := DumpAWSRegion(ctx, cfgCopy)
			resultLock.Lock()
			if err != nil {
				errors = append(errors, err)
			}
			result.Regions[regionName] = regionDump
			resultLock.Unlock()
		})
	}

	err = executor.Wait(ctx)
	if err != nil {
		return result, err
	}
	if len(errors) > 0 {
		return result, common.NewErrors(errors)
	}

	return result, err
}

func DumpAWSRegion(ctx context.Context, cfg aws.Config) (AWSRegion, error) {
	result := NewAWSRegion(cfg.Region)
	errorsCh := make(chan error)
	executor := executor.NewExecutor(0)

	fetchEC2(ctx, cfg, executor, errorsCh, &result)
	fetchELBs(ctx, cfg, executor, errorsCh, &result)
	fetchS3(ctx, cfg, executor, errorsCh, &result)
	fetchOpsworks(ctx, cfg, executor, errorsCh, &result)

	errors := make([]error, 0)
	consume := true
	for consume {
		select {
		case <-executor.Done():
			consume = false
		case err := <-errorsCh:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		errStrings := make([]string, len(errors))
		log.Printf("The following errors occured, cannot continue:")
		for i, err := range errors {
			log.Printf(" - %v", err)
			errStrings[i] = err.Error()
		}
		return result, fmt.Errorf("multiple errors [%s]", strings.Join(errStrings, ", "))
	}

	return result, nil
}

func fetchEC2(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *AWSRegion) {
	executor.Launch(ctx, func() {
		reservations, err := loader.FetchAllInstances(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EC2 instances: %w", err)
		}
		result.EC2.Reservations = reservations
	})

	executor.Launch(ctx, func() {
		volumes, err := loader.FetchAllEBSVolumes(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EBS volumes: %w", err)
		}
		result.EC2.Volumes = volumes
	})
}

func fetchS3(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *AWSRegion) {
	executor.Launch(ctx, func() {
		buckets, err := loader.FetchAllBuckets(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EC2 instances: %w", err)
		}
		result.S3.Buckets = buckets
	})

	// executor.Launch(ctx, func() {
	// 	<-bucketsDoneCh
	// 	var lock sync.Mutex
	// 	for _, bucket := range result.S3.Buckets {
	// 		bucketName := *bucket.Name
	// 		executor.Launch(ctx, func() {
	// 			tags, err := loader.FetchBucketTags(ctx, cfg, bucketName)
	// 			if err != nil {
	// 				errorsCh <- fmt.Errorf("error while fetching tags for S3 bucket %s: %w", bucketName, err)
	// 			}
	// 			lock.Lock()
	// 			defer lock.Unlock()
	// 			result.S3.BucketTags[bucketName] = tags
	// 		})
	// 	}
	// })

}

func fetchELBs(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *AWSRegion) {
	executor.Launch(ctx, func() {
		elbs, err := loader.FetchAllV1ELBs(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all ELBs (v1): %w", err)
		}
		result.ELB.V1.LoadBalancers = elbs
	})

	executor.Launch(ctx, func() {
		elbs, err := loader.FetchAllV2ELBs(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all ELBs (v2): %w", err)
		}
		result.ELB.V2.LoadBalancers = elbs
	})
}

func fetchOpsworks(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *AWSRegion) {
	stacksDoneCh := executor.Launch(ctx, func() {
		stacks, err := loader.FetchAllOpsworksStacks(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all Opsworks stacks: %w", err)
		}
		result.Opsworks.Stacks = stacks
	})

	executor.Launch(ctx, func() {
		<-stacksDoneCh
		var layersLock sync.Mutex
		var appsLock sync.Mutex
		var instancesLock sync.Mutex
		for _, stack := range result.Opsworks.Stacks {
			stackId := *stack.StackId

			executor.Launch(ctx, func() {
				layers, err := loader.FetchAllOpsworksLayers(ctx, cfg, stackId)
				if err != nil {
					errorsCh <- fmt.Errorf("error while fetching all Opsworks layers for stack %s: %w", stackId, err)
				}
				layersLock.Lock()
				defer layersLock.Unlock()
				result.Opsworks.Layers = append(result.Opsworks.Layers, layers...)
			})

			executor.Launch(ctx, func() {
				apps, err := loader.FetchAllOpsworksApps(ctx, cfg, stackId)
				if err != nil {
					errorsCh <- fmt.Errorf("error while fetching all Opsworks apps for stack %s: %w", stackId, err)
				}
				appsLock.Lock()
				defer appsLock.Unlock()
				result.Opsworks.Apps = append(result.Opsworks.Apps, apps...)
			})

			executor.Launch(ctx, func() {
				instances, err := loader.FetchAllOpsworksInstances(ctx, cfg, stackId)
				if err != nil {
					errorsCh <- fmt.Errorf("error while fetching all Opsworks instances for stack %s: %w", stackId, err)
				}
				instancesLock.Lock()
				defer instancesLock.Unlock()
				result.Opsworks.Instances = append(result.Opsworks.Instances, instances...)
			})
		}
	})
}
