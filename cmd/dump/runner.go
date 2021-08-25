package main

import (
	"aws-tools/aws/ec2"
	"aws-tools/aws/elb"
	"aws-tools/aws/opsworks"
	"aws-tools/executor"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func run(ctx context.Context, cfg aws.Config) error {
	maxParallelism := 10

	result := Result{
		EC2: EC2{},
		ELB: ELB{
			V1: ELBv1{},
			V2: ELBv2{},
		},
		Opsworks: Opsworks{},
	}
	errorsCh := make(chan error)
	executor := executor.NewExecutor(maxParallelism)

	fetchEC2(ctx, cfg, executor, errorsCh, &result)
	fetchELBs(ctx, cfg, executor, errorsCh, &result)
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
		return fmt.Errorf("multiple errors [%s]", strings.Join(errStrings, ", "))
	}

	log.Print("Data fully loaded, encoding to json")

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error while encoding result to json: %v", err)
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}

func fetchEC2(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *Result) {
	executor.Launch(ctx, func() {
		reservations, err := ec2.FetchAllInstances(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EC2 instances: %w", err)
		}
		result.EC2.Reservations = reservations
	})

	executor.Launch(ctx, func() {
		volumes, err := ec2.FetchAllEBSVolumes(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EBS volumes: %w", err)
		}
		result.EC2.Volumes = volumes
	})
}

func fetchELBs(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *Result) {
	executor.Launch(ctx, func() {
		elbs, err := elb.FetchAllV1ELBs(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all ELBs (v1): %w", err)
		}
		result.ELB.V1.LoadBalancers = elbs
	})

	executor.Launch(ctx, func() {
		elbs, err := elb.FetchAllV2ELBs(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all ELBs (v2): %w", err)
		}
		result.ELB.V2.LoadBalancers = elbs
	})
}

func fetchOpsworks(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *Result) {
	stacksDoneCh := executor.Launch(ctx, func() {
		stacks, err := opsworks.FetchAllOpsworksStacks(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all Opsworks stacks: %w", err)
		}
		result.Opsworks.Stacks = stacks
	})

	<-stacksDoneCh
	var layersLock sync.Mutex
	var appsLock sync.Mutex
	var instancesLock sync.Mutex
	for _, stack := range result.Opsworks.Stacks {
		stackId := *stack.StackId

		executor.Launch(ctx, func() {
			layers, err := opsworks.FetchAllOpsworksLayers(ctx, cfg, stackId)
			if err != nil {
				errorsCh <- fmt.Errorf("error while fetching all Opsworks layers for stack %s: %w", stackId, err)
			}
			layersLock.Lock()
			defer layersLock.Unlock()
			result.Opsworks.Layers = append(result.Opsworks.Layers, layers...)
		})

		executor.Launch(ctx, func() {
			apps, err := opsworks.FetchAllOpsworksApps(ctx, cfg, stackId)
			if err != nil {
				errorsCh <- fmt.Errorf("error while fetching all Opsworks apps for stack %s: %w", stackId, err)
			}
			appsLock.Lock()
			defer appsLock.Unlock()
			result.Opsworks.Apps = append(result.Opsworks.Apps, apps...)
		})

		executor.Launch(ctx, func() {
			instances, err := opsworks.FetchAllOpsworksInstances(ctx, cfg, stackId)
			if err != nil {
				errorsCh <- fmt.Errorf("error while fetching all Opsworks instances for stack %s: %w", stackId, err)
			}
			instancesLock.Lock()
			defer instancesLock.Unlock()
			result.Opsworks.Instances = append(result.Opsworks.Instances, instances...)
		})
	}
}
