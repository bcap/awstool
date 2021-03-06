package resolve

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	awst "awstool/aws"
	"awstool/aws/ec2"
	"awstool/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
)

type printOptions struct {
	publicIp  bool
	privateIp bool
	tags      []string
	allTags   bool
	urlEncode bool
	header    bool
	template  *template.Template
}

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "resolve",
		Short:         "resolves ec2 instances by a given set of inputs",
		SilenceErrors: true,
	}

	var instanceId string
	var tags []string

	printOptions := printOptions{}
	var noURLEncode bool
	var templ string

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
		&printOptions.tags, "print-tags", "T", []string{},
		"By default the command only prints the Name tag. Pass a list of tags keys that "+
			"should be printed instead. See also --print-all-tags",
	)

	cmd.Flags().BoolVarP(
		&printOptions.allTags, "print-all-tags", "A", false,
		"Also prints all tags associated to the instances. Overrides --print-tags",
	)

	cmd.Flags().BoolVarP(
		&printOptions.publicIp, "public", "u", false,
		"Only print the public ip of the instance, if it has one",
	)

	cmd.Flags().BoolVarP(
		&printOptions.privateIp, "private", "r", false,
		"Only print the private ip of the instance, if it has one",
	)

	cmd.Flags().BoolVarP(
		&noURLEncode, "no-url-encode", "E", false,
		"By default when printing tags their values are URL encoded to avoid whitespacing issues. "+
			"Use this flag to avoid such mechanism",
	)

	cmd.Flags().BoolVarP(
		&printOptions.header, "header", "H", false,
		"Also print a header on the first line, which will name the columns being printed",
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

		if printOptions.publicIp && printOptions.privateIp {
			return fmt.Errorf("cannot have both private and public ip print options enabled")
		}

		printOptions.urlEncode = !noURLEncode

		if templ != "" {
			var err error
			printOptions.template, err = template.New("user_input").Parse(templ)
			if err != nil {
				return fmt.Errorf("invalid template %q: %w", templ, err)
			}
		}

		// We silence usage here instead of setting in the command struct declaration because it is
		// only at this point forward that we want to not display the usage when an error occurs,
		// as it will be an execution error, not a parsing/usage error
		// See more at https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		resolution, err := resolve(cmd.Context(), **awsCfg, instanceId, parsedTags)
		if err != nil {
			return fmt.Errorf("failed while fetching instances: %w", err)
		}
		printInstances(resolution, printOptions)
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

func printInstances(aws *awst.AWS, printOptions printOptions) {
	printHeader(printOptions)
	for _, region := range aws.Regions {
		for _, reservation := range region.EC2.Reservations {
			for _, instance := range reservation.Instances {
				printInstance(region.Region, reservation, instance, printOptions)
			}
		}
	}
}

func printHeader(printOptions printOptions) {
	if !printOptions.header || printOptions.template != nil {
		return
	}
	if printOptions.privateIp {
		fmt.Println("#privateIp")
	} else if printOptions.publicIp {
		fmt.Println("#privateIp")
	} else {
		fmt.Print("#region #instanceid #privateIp #publicIp ")
		if printOptions.allTags || len(printOptions.tags) > 0 {
			fmt.Println("#tags")
		} else {
			fmt.Println("#name")
		}
	}
}

type templateData struct {
	Region      string
	Reservation *ec2Types.Reservation
	Instance    *ec2Types.Instance
}

func printInstance(region string, reservation ec2Types.Reservation, instance ec2Types.Instance, printOptions printOptions) {
	if printOptions.template != nil {
		buf := strings.Builder{}
		data := templateData{
			Region:      region,
			Reservation: &reservation,
			Instance:    &instance,
		}
		if err := printOptions.template.Execute(&buf, data); err != nil {
			fmt.Printf("template execution error: %v\n", err)
		}
		fmt.Println(buf.String())
		return
	}

	if printOptions.privateIp {
		if instance.PrivateIpAddress != nil {
			fmt.Println(safeString(instance.PrivateIpAddress))
		}
		return
	}

	if printOptions.publicIp {
		if instance.PublicIpAddress != nil {
			fmt.Println(safeString(instance.PublicIpAddress))
		}
		return
	}

	fmt.Printf(
		"%s %s %s %s %s\n",
		region,
		safeString(instance.InstanceId),
		safeString(instance.PrivateIpAddress),
		safeString(instance.PublicIpAddress),
		tagsString(&instance, printOptions),
	)
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

func tagsString(instance *ec2Types.Instance, printOptions printOptions) string {
	// if no tags were passed and print-all-tags is not enabled, just return the name
	if len(printOptions.tags) == 0 && !printOptions.allTags {
		name := ""
		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				name = *tag.Value
				break
			}
		}
		if printOptions.urlEncode {
			name = url.PathEscape(name)
		}
		return safeString(&name)
	}

	result := []string{}
	appendTag := func(tag ec2Types.Tag) {
		value := *tag.Value
		if printOptions.urlEncode {
			value = url.PathEscape(value)
		}
		result = append(result, *tag.Key+":"+value)
	}

	if printOptions.allTags {
		for _, tag := range instance.Tags {
			appendTag(tag)
		}

	} else {
		tagMap := map[string]ec2Types.Tag{}
		for _, tag := range instance.Tags {
			tagMap[*tag.Key] = tag
		}
		for _, tagToFind := range printOptions.tags {
			if tag, ok := tagMap[tagToFind]; ok {
				appendTag(tag)
			}
		}
	}

	return strings.Join(result, ",")
}
