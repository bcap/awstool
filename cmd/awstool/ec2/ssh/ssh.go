package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"awstool/aws/ec2"
	"awstool/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type identifier struct {
	user     string
	instance string
}

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "ssh",
		Short:         "connects to an instance via ssh",
		SilenceErrors: true,
	}

	cmd.Args = cobra.ExactArgs(1)

	var publicIp bool

	cmd.Flags().BoolVarP(
		&publicIp, "public-ip", "P", false,
		"By default the command sshs through the private ip of the instance. "+
			"Set this to use the public ip instead",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		identifier, err := parseIdentifier(args[0])
		if err != nil {
			return err
		}

		// We silence usage here instead of setting in the command struct declaration because it is
		// only at this point forward that we want to not display the usage when an error occurs,
		// as it will be an execution error, not a parsing/usage error
		// See more at https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		region, instance, err := resolve(cmd.Context(), **awsCfg, identifier.instance)
		if err != nil {
			return fmt.Errorf("could not resolve instance with id %s: %w", identifier.instance, err)
		}

		if identifier.user == "" {
			user, err := discoverUser(cmd.Context(), **awsCfg, *instance.ImageId)
			if err != nil {
				return fmt.Errorf("failed to discover user for image %s: %w", *instance.ImageId, err)
			}
			if user == "" {
				log.Warnf(
					"Could not discover user for image %s, proceeding with ssh defaults",
					*instance.ImageId,
				)
			}
			identifier.user = user
		} else {
			log.Debugf("User %s passed in, so not trying to discover the user automatically", identifier.user)
		}

		if err = ssh(cmd.Context(), identifier.user, region, instance, publicIp); err != nil {
			exitErr := &exec.ExitError{}
			if errors.As(err, &exitErr) {
				return fmt.Errorf("ssh finished with exit code %d", exitErr.ExitCode())
			}
			return err
		}
		return nil
	}

	return &cmd
}

func resolve(ctx context.Context, cfg aws.Config, instanceId string) (string, *ec2Types.Instance, error) {
	result, err := loader.LoadAWS(
		ctx, cfg,
		loader.WithServices("ec2"),
		loader.WithEC2FetchOptions(ec2.WithInstanceIds(instanceId)),
	)
	if err != nil {
		return "", nil, err
	}
	instances := []ec2Types.Instance{}
	instanceRegion := map[string]string{}
	for _, region := range result.Regions {
		for _, reservation := range region.EC2.Reservations {
			for _, instance := range reservation.Instances {
				instances = append(instances, instance)
				instanceRegion[*instance.InstanceId] = region.Region
			}
		}
	}
	if len(instances) == 0 {
		return "", nil, fmt.Errorf("could not find instance %q", instanceId)
	}
	if len(instances) > 1 {
		instancesMsg := make([]string, len(instances))
		for idx, instance := range instances {
			instancesMsg[idx] = *instance.InstanceId + " in " + instanceRegion[*instance.InstanceId]
		}
		return "", nil, fmt.Errorf("found multiple instances with id %q: %v. This is unusual and should not happen", instanceId, instancesMsg)
	}
	return instanceRegion[*instances[0].InstanceId], &instances[0], nil
}

func ssh(ctx context.Context, user string, region string, instance *ec2Types.Instance, usePublicIp bool) error {
	address := *instance.PrivateIpAddress
	if usePublicIp {
		address = *instance.PublicIpAddress
	}
	userMsg := "user " + user
	identifier := user + "@" + address
	if user == "" {
		userMsg = "no user set"
		identifier = address
	}

	fmt.Fprintf(
		os.Stderr,
		"> Connecting to %s in %s with %s (ssh %s)\n",
		*instance.InstanceId,
		region,
		userMsg,
		identifier,
	)

	execution := exec.CommandContext(ctx, "ssh", identifier)
	execution.Stdin = os.Stdin
	execution.Stderr = os.Stderr
	execution.Stdout = os.Stdout
	err := execution.Start()
	if err != nil {
		return fmt.Errorf("failed to launch ssh: %w", err)
	}
	return execution.Wait()
}

func discoverUser(ctx context.Context, cfg aws.Config, imageId string) (string, error) {
	image, err := ec2.GetImage(ctx, cfg, imageId)
	if err != nil {
		return "", err
	}
	if image == nil {
		return "", nil
	}

	// The following is a naive implementation of what is described on
	//    https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/connection-prereqs.html
	// Available image locations can be dumped with `awsv ec2 describe-images`
	os2UserMap := map[string]string{
		"amazon": "ec2-user",
		"centos": "ec2-user",
		"debian": "admin",
		"fedora": "ec2-user",
		"suse":   "ec2-user",
		"ubuntu": "ubuntu",
	}

	// to minimize false positives, we check in this order
	orderToTry := []string{
		"ubuntu",
		"centos",
		"fedora",
		"debian",
		"suse",
		"amazon",
	}

	location := strings.ToLower(*image.ImageLocation)
	for _, os := range orderToTry {
		if strings.Contains(location, os) {
			user := os2UserMap[os]
			log.Debugf(
				"image %s has location %s, which seems to be an %s image. Moving on with user %s",
				imageId,
				location,
				os,
				user,
			)
			return user, nil
		}
	}
	return "", nil
}

func parseIdentifier(idStr string) (identifier, error) {
	idx := strings.Index(idStr, "@")
	if idx == 0 {
		return identifier{}, fmt.Errorf("invalid identifier %q: user must value a value or not be set", idStr)
	}
	result := identifier{instance: idStr}
	if idx > 0 {
		result.user = idStr[:idx]
		result.instance = idStr[idx+1:]
	}
	return result, nil
}
