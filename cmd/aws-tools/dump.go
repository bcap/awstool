package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func DumpCommand() *cobra.Command {
	var regions []string
	var excludeRegions []string
	var services []string
	var excludeServices []string
	var listServicesOnly bool

	cmd := cobra.Command{
		Use:           "dump",
		Short:         "generates a single json dumping the results of many different description APIs from AWS",
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringSliceVarP(
		&regions, "regions", "r", []string{},
		"Dump data for only those regions. If not specified, all regions will be dumped. "+
			"See also --exclude-regions",
	)
	cmd.PersistentFlags().StringSliceVarP(
		&excludeRegions, "exclude-regions", "R", []string{},
		"Do not dump data for those regions. This takes precedence over --regions",
	)

	cmd.PersistentFlags().StringSliceVarP(
		&services, "services", "s", []string{},
		"Dump only those services. If not specified, all implemented services will be dumped. "+
			"See also --list-services and --exclude-services",
	)
	cmd.PersistentFlags().StringSliceVarP(
		&excludeServices, "exclude-services", "S", []string{},
		"Do not dump data for those services. This takes precedence over --services. See also --list-services",
	)
	cmd.PersistentFlags().BoolVar(
		&listServicesOnly, "list-services", false,
		"List which services can be dumped and exit. Those are inputs for the --services and --exclude-services flags",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		loaderOptions := []loader.Option{
			loader.WithRegions(regions...),
			loader.WithoutRegions(excludeRegions...),
			loader.WithServices(services...),
			loader.WithoutServices(excludeServices...),
		}

		if listServicesOnly {
			listServices()
			return nil
		}

		// we silence usage here instead of setting in the command struct declaration because it is
		// only at this point forward that we want to not display the usage when an error occurs,
		// as it will be an execution error, not a parsing/usage error
		// see more at https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		return dump(cmd.Context(), awsCfg, loaderOptions...)
	}

	return &cmd
}

func listServices() {
	for _, service := range loader.ListServices() {
		fmt.Println(service)
	}
}

func dump(ctx context.Context, cfg aws.Config, options ...loader.Option) error {
	result, err := loader.LoadAWS(ctx, cfg, options...)
	if err != nil {
		log.Errorf("Error while loading data: %v", err)
		return err
	}

	log.Info("Data fully loaded, encoding to json")
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		log.Errorf("Error while encoding result to json: %v", err)
		return err
	}

	// using raw os.Stdout.Write([]byte) to avoid copying potentially huge json byte
	// data to string
	os.Stdout.Write(jsonBytes)
	os.Stdout.Write([]byte{'\n'})

	return nil
}
