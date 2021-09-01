package dump

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"aws-tools/common"
	"aws-tools/executor"
	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
)

//
// Options for the dumper
//

type options struct {
	includeRegions map[string]struct{}
	excludeRegions map[string]struct{}

	includeServices map[string]struct{}
	excludeServices map[string]struct{}
}

type Option func(opts *options)

func WithRegions(regions []string) Option {
	return func(opts *options) {
		for _, region := range regions {
			opts.includeRegions[strings.ToLower(region)] = struct{}{}
		}
	}
}

func WithoutRegions(regions []string) Option {
	return func(opts *options) {
		for _, region := range regions {
			opts.excludeRegions[strings.ToLower(region)] = struct{}{}
		}
	}
}

func WithServices(services []string) Option {
	return func(opts *options) {
		for _, service := range services {
			opts.includeServices[strings.ToLower(service)] = struct{}{}
		}
	}
}

func WithoutServices(services []string) Option {
	return func(opts *options) {
		for _, service := range services {
			opts.excludeServices[strings.ToLower(service)] = struct{}{}
		}
	}
}

func newOptions(fns []Option) options {
	options := options{
		includeRegions:  map[string]struct{}{},
		excludeRegions:  map[string]struct{}{},
		includeServices: map[string]struct{}{},
		excludeServices: map[string]struct{}{},
	}
	for _, fn := range fns {
		fn(&options)
	}
	return options
}

//
// Dumper functions
//

func DumpAWS(ctx context.Context, cfg aws.Config, options ...Option) (AWS, error) {
	opts := newOptions(options)
	result := NewAWS()
	errorsCh := make(chan error)
	var resultLock sync.Mutex

	regions, err := getRegions(ctx, cfg, opts)
	if err != nil {
		return result, err
	}

	executor := executor.NewExecutor(0)
	for _, region := range regions {
		regionRef := region
		executor.Launch(ctx, func() {
			regionDump, err := DumpAWSRegion(ctx, cfg, regionRef, options...)
			if err != nil {
				errorsCh <- err
				return
			}
			resultLock.Lock()
			result.Regions[regionRef] = regionDump
			resultLock.Unlock()
		})
	}

	// global services
	if shouldFetchService("iam", opts) {
		fetchIAM(ctx, cfg, executor, errorsCh, &result)
	}

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
		return result, common.NewErrors(errors)
	}

	return result, err
}

func DumpAWSRegion(ctx context.Context, cfg aws.Config, region string, options ...Option) (AWSRegion, error) {
	cfg.Region = region

	servicesCfg := map[string]func(context.Context, aws.Config, *executor.Executor, chan<- error, *AWSRegion){
		"ec2":              fetchEC2,
		"elb":              fetchELBs,
		"s3":               fetchS3,
		"opsworks":         fetchOpsworks,
		"elasticbeanstalk": fetchElasticBeanstalk,
	}

	opts := newOptions(options)
	result := NewAWSRegion(region)
	errorsCh := make(chan error)
	executor := executor.NewExecutor(0)
	for svc, fn := range servicesCfg {
		if shouldFetchService(svc, opts) {
			fn(ctx, cfg, executor, errorsCh, &result)
		}
	}

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
		return result, common.NewErrors(errors)
	}

	return result, nil
}

func fetchIAM(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *AWS) {
	usersDoneCh := executor.Launch(ctx, func() {
		users, err := loader.FetchAllUsers(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all IAM users: %w", err)
		}
		result.IAM.Users = users
	})

	executor.Launch(ctx, func() {
		roles, err := loader.FetchAllRoles(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all IAM roles: %w", err)
		}
		result.IAM.Roles = roles
	})

	executor.Launch(ctx, func() {
		groups, err := loader.FetchAllGroups(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all IAM groups: %w", err)
		}
		result.IAM.Groups = groups
	})

	executor.Launch(ctx, func() {
		policies, err := loader.FetchAllPolicies(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all IAM policies: %w", err)
		}
		result.IAM.Policies = policies
	})

	executor.Launch(ctx, func() {
		<-usersDoneCh
		var lock sync.Mutex
		for _, user := range result.IAM.Users {
			username := *user.UserName
			executor.Launch(ctx, func() {
				accessKeys, err := loader.FetchAllAccessKeys(ctx, cfg, username)
				if err != nil {
					errorsCh <- fmt.Errorf("error while fetching all IAM access keys for %s: %w", username, err)
				}
				lock.Lock()
				result.IAM.AccessKeys[username] = accessKeys
				lock.Unlock()
			})
		}
	})

	executor.Launch(ctx, func() {
		<-usersDoneCh
		var lock sync.Mutex
		for _, user := range result.IAM.Users {
			username := *user.UserName
			executor.Launch(ctx, func() {
				groups, err := loader.FetchAllUserGroups(ctx, cfg, username)
				if err != nil {
					errorsCh <- fmt.Errorf("error while fetching all IAM user groups for %s: %w", username, err)
				}
				lock.Lock()
				result.IAM.UserGroups[username] = groups
				lock.Unlock()
			})
		}
	})
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
	bucketsDoneCh := executor.Launch(ctx, func() {
		buckets, err := loader.FetchAllBuckets(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all EC2 instances: %w", err)
		}
		result.S3.Buckets = buckets
	})

	executor.Launch(ctx, func() {
		<-bucketsDoneCh
		// var lock sync.Mutex
		// for _, bucket := range result.S3.Buckets {
		// 	bucketName := *bucket.Name
		// 	executor.Launch(ctx, func() {
		// 		tags, err := loader.FetchBucketTags(ctx, cfg, bucketName)
		// 		if err != nil {
		// 			errorsCh <- fmt.Errorf("error while fetching tags for S3 bucket %s: %w", bucketName, err)
		// 		}
		// 		lock.Lock()
		// 		defer lock.Unlock()
		// 		result.S3.BucketTags[bucketName] = tags
		// 	})
		// }
	})

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

func fetchElasticBeanstalk(ctx context.Context, cfg aws.Config, executor *executor.Executor, errorsCh chan<- error, result *AWSRegion) {
	executor.Launch(ctx, func() {
		apps, err := loader.FetchAllElasticBeanstalkApplications(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all Opsworks stacks: %w", err)
		}
		result.ElasticBeanstalk.Applications = apps
	})

	executor.Launch(ctx, func() {
		envs, err := loader.FetchAllElasticBeanstalkEnvironments(ctx, cfg)
		if err != nil {
			errorsCh <- fmt.Errorf("error while fetching all Opsworks stacks: %w", err)
		}
		result.ElasticBeanstalk.Environments = envs
	})
}

func getRegions(ctx context.Context, cfg aws.Config, options options) ([]string, error) {
	// If no regions were passed in, them include all
	regions := options.includeRegions
	if len(regions) == 0 {
		regionObjs, err := loader.FetchAllRegions(ctx, cfg)
		if err != nil {
			return nil, err
		}
		for _, regionObj := range regionObjs {
			regions[*regionObj.RegionName] = struct{}{}
		}
	}

	// Remove exclusions
	for region, _ := range options.excludeRegions {
		delete(regions, region)
	}

	// convert to sorted string slice
	result := make([]string, 0, len(regions))
	for region, _ := range regions {
		result = append(result, region)
	}
	sort.Strings(result)

	return result, nil
}

func shouldFetchService(service string, options options) bool {
	service = strings.ToLower(service)
	_, excluded := options.excludeServices[service]
	if excluded {
		return false
	}
	// if no explicit inclusions were done then we want all services
	if len(options.includeServices) == 0 {
		return true
	}
	_, included := options.includeServices[service]
	return included
}
