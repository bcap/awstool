package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	awst "aws-tools/aws"
	"aws-tools/aws/ec2"
	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
)

func ResolveCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:           "resolve",
		Short:         "resolves ec2 instances by a given set of inputs",
		SilenceErrors: true,
	}

	var instanceId string
	var tags []string
	var printTags []string
	var publicIp bool
	var privateIp bool
	var templ string
	var noURLEncode bool

	cmd.Flags().StringVarP(
		&instanceId, "instance-id", "i", "",
		"Find the instance by its id",
	)

	cmd.Flags().StringSliceVarP(
		&tags, "tags", "t", []string{},
		"Find the instance by tag key/value pairs. Values are ANDed togther. "+
			"Eg: --tags Owner:Bruno,Env:development. Alternatively: --tags Owner:Bruno --tags Env:development",
	)

	cmd.Flags().StringSliceVarP(
		&printTags, "print-tags", "T", []string{},
		"By default the command only prints the Name tag. Pass a list of tags keys that should be printed instead",
	)

	cmd.Flags().BoolVarP(
		&publicIp, "public", "u", false,
		"Only print the public ip of the instance, if it has one",
	)

	cmd.Flags().BoolVarP(
		&privateIp, "private", "r", false,
		"Only print the private ip of the instance, if it has one",
	)

	cmd.Flags().BoolVarP(
		&noURLEncode, "no-url-encode", "E", false,
		"By default when printing tags their values are URL encoded to avoid whitespacing issues. "+
			"Use this flag to avoid such mechanism",
	)

	cmd.Flags().StringVar(
		&templ, "template", "",
		"Print using a golang template instead. Template syntax is defined at https://pkg.go.dev/text/template. "+
			"A simple template for printing region and instance id: '{{.Region}} {{.Instance.InstanceId}}'. "+
			"The struct passed to the template engine can be checked by reading the source code "+
			"for this command. Instead you can also inspect and navigate structs available for a template engine "+
			"by using the template '{{printf \"%#+v\" .}}' and navigating from that point forward, for instance with "+
			"'{{printf \"%#+v\" .Instance}}' and so on",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		parsedTags, err := parseTags(tags)
		if err != nil {
			return err
		}

		if publicIp && privateIp {
			return fmt.Errorf("cannot have both private and public ip print options enabled")
		}

		var templEngine *template.Template
		if templ != "" {
			var err error
			templEngine, err = template.New("").Parse(templ)
			if err != nil {
				return fmt.Errorf("invalid template %q: %w", templ, err)
			}
		}

		// We silence usage here instead of setting in the command struct declaration because it is
		// only at this point forward that we want to not display the usage when an error occurs,
		// as it will be an execution error, not a parsing/usage error
		// See more at https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		resolution, err := resolve(cmd.Context(), awsCfg, instanceId, parsedTags)
		if err != nil {
			return fmt.Errorf("failed while fetching instances: %w", err)
		}
		printInstances(resolution, printTags, publicIp, privateIp, !noURLEncode, templEngine)
		return nil
	}

	return &cmd
}

func resolve(ctx context.Context, cfg aws.Config, instanceId string, tags map[string]string) (*awst.AWS, error) {
	fetchOpts := []ec2.FetchOption{}
	if instanceId != "" {
		fetchOpts = append(fetchOpts, ec2.WithInstanceIds(instanceId))
	}
	for k, v := range tags {
		fetchOpts = append(fetchOpts, ec2.WithTag(k, v))
	}
	result, err := loader.LoadAWS(
		ctx, cfg,
		loader.WithServices("ec2"),
		loader.WithEC2FetchOptions(fetchOpts...),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type templateHelper struct {
	Region      string
	Reservation *ec2Types.Reservation
	Instance    *ec2Types.Instance
}

func printInstances(aws *awst.AWS, tags []string, publicIpOnly bool, privateIpOnly bool, urlEncode bool, template *template.Template) {
	for _, region := range aws.Regions {
		for _, reservation := range region.EC2.Reservations {
			for _, instance := range reservation.Instances {
				if template != nil {
					buf := strings.Builder{}
					helper := templateHelper{Region: region.Region, Reservation: &reservation, Instance: &instance}
					if err := template.Execute(&buf, helper); err != nil {
						fmt.Printf("template execution error: %v\n", err)
					}
					fmt.Println(buf.String())
					continue
				}

				if privateIpOnly {
					if instance.PrivateIpAddress != nil {
						fmt.Println(safeString(instance.PrivateIpAddress))
					}
					continue
				}

				if publicIpOnly {
					if instance.PublicIpAddress != nil {
						fmt.Println(safeString(instance.PublicIpAddress))
					}
					continue
				}

				fmt.Printf(
					"%s %s %s %s %s\n",
					region.Region,
					safeString(instance.InstanceId),
					safeString(instance.PrivateIpAddress),
					safeString(instance.PublicIpAddress),
					tagsString(&instance, tags, urlEncode),
				)
			}
		}
	}
}

func safeString(s *string) string {
	if s == nil || *s == "" {
		return "<N/A>"
	}
	return *s
}

func parseTags(tags []string) (map[string]string, error) {
	parsedTags := map[string]string{}
	if len(tags) == 0 {
		return parsedTags, nil
	}
	for _, kvpair := range tags {
		kvpair = strings.TrimSpace(kvpair)
		separatorIdx := strings.Index(kvpair, ":")
		if separatorIdx == -1 {
			return nil, fmt.Errorf("invalid tags specification in %q: cannot parse %q: missing \":\" separator", tags, kvpair)
		}
		key := kvpair[0:separatorIdx]
		value := kvpair[separatorIdx+1:]
		if len(key) == 0 {
			return nil, fmt.Errorf("invalid tags specification in %q: cannot parse %q: no key", tags, kvpair)
		}
		if len(value) == 0 {
			return nil, fmt.Errorf("invalid tags specification in %q: cannot parse %q: no value", tags, kvpair)
		}
		parsedTags[key] = value
	}
	return parsedTags, nil
}

func tagsString(instance *ec2Types.Instance, tags []string, urlEncode bool) string {
	// if no tags were passed, just return the name
	if tags == nil || len(tags) == 0 {
		name := ""
		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				name = *tag.Value
				break
			}
		}
		if urlEncode {
			name = url.PathEscape(name)
		}
		return safeString(&name)
	}

	result := []string{}
	// stupid O(n^2) algorithm, but dataset is small so we dont care much
	for _, tagToFind := range tags {
		for _, tag := range instance.Tags {
			if *tag.Key == tagToFind {
				value := *tag.Value
				if urlEncode {
					value = url.PathEscape(value)
				}
				result = append(result, *tag.Key+":"+value)
			}
		}
	}
	return strings.Join(result, ",")
}
