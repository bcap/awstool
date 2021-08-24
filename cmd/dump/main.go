package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"aws-tools/ec2"
	"aws-tools/elb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := loadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config, cannot continue: %v", err)
		os.Exit(1)
	}

	if err := run(ctx, cfg); err != nil {
		os.Exit(1)
	}

	log.Println("Done")
}

func loadConfig(ctx context.Context) (aws.Config, error) {
	log.Println("Loading default config")

	opts := []func(*config.LoadOptions) error{}
	//TODO gate this behind a flag
	opts = append(opts, config.WithSharedConfigProfile("vtex"))

	return config.LoadDefaultConfig(ctx, opts...)
}

func run(ctx context.Context, cfg aws.Config) error {
	var wg sync.WaitGroup
	done := make(chan struct{})

	launch := func(fn func()) {
		wg.Add(1)
		go func() {
			fn()
			wg.Done()
		}()
	}

	result := Result{
		EC2: EC2{},
		ELB: ELB{
			V1: ELBv1{},
			V2: ELBv2{},
		},
	}
	errorsCh := make(chan error)

	launch(func() {
		reservations, err := ec2.FetchAllInstances(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EC2 instances: %w", err)
		}
		result.EC2.Reservations = reservations
	})

	launch(func() {
		volumes, err := ec2.FetchAllEBSVolumes(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EBS volumes: %w", err)
		}
		result.EC2.Volumes = volumes
	})

	launch(func() {
		elbs, err := elb.FetchAllV1ELBs(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all ELBs (v1): %w", err)
		}
		result.ELB.V1.LoadBalancers = elbs
	})

	launch(func() {
		elbs, err := elb.FetchAllV2ELBs(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all ELBs (v2): %w", err)
		}
		result.ELB.V2.LoadBalancers = elbs
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	errors := make([]error, 0)
	consume := true
	for consume {
		select {
		case <-done:
			consume = false
		case err := <-errorsCh:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		errStrings := make([]string, len(errors))
		log.Fatalf("The following errors occured, cannot continue:")
		for i, err := range errors {
			log.Fatalf(" - %v", err)
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
