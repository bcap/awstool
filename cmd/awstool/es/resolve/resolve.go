package resolve

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	awst "awstool/aws"
	"awstool/aws/elasticsearch"
	"awstool/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	esTypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/spf13/cobra"
)

type printOptions struct {
	header   bool
	template *template.Template
}

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "resolve",
		Short:         "resolves elasticsearch domains by a given set of inputs",
		SilenceErrors: true,
	}

	var domain string

	printOptions := printOptions{}
	var templ string

	cmd.Flags().StringVarP(
		&domain, "domain", "d", "",
		"Find the domain by its name",
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

		resolution, err := resolve(cmd.Context(), **awsCfg, domain)
		if err != nil {
			return fmt.Errorf("failed while fetching instances: %w", err)
		}
		printDomains(resolution, printOptions)
		return nil
	}

	return &cmd
}

func resolve(ctx context.Context, cfg aws.Config, domain string) (*awst.AWS, error) {
	fetchOpts := []elasticsearch.FetchOption{}
	if domain != "" {
		fetchOpts = append(fetchOpts, elasticsearch.WithDomains(domain))
	}
	result, err := loader.LoadAWS(
		ctx, cfg,
		loader.WithServices("elasticsearch"),
		loader.WithESFetchOptions(fetchOpts...),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func printDomains(aws *awst.AWS, printOptions printOptions) {
	printHeader(printOptions)
	for _, region := range aws.Regions {
		for _, domain := range region.Elasticsearch.Domains {
			printDomain(region.Region, domain, printOptions)
		}
	}
}

func printHeader(printOptions printOptions) {
	if !printOptions.header || printOptions.template != nil {
		return
	}
	fmt.Println("#region #domain #arn #endpoints")
}

func printDomain(region string, domain *awst.ElasticsearchDomain, printOptions printOptions) {
	if printOptions.template != nil {
		buf := strings.Builder{}
		if err := printOptions.template.Execute(&buf, domain); err != nil {
			fmt.Printf("template execution error: %v\n", err)
		}
		fmt.Println(buf.String())
		return
	}
	fmt.Printf(
		"%s %s %s %s\n",
		region,
		*domain.Status.DomainName,
		*domain.Status.ARN,
		strings.Join(endpoints(domain.Status), ","),
	)
}

func endpoints(domain *esTypes.ElasticsearchDomainStatus) []string {
	result := []string{}
	if domain.Endpoint != nil {
		result = append(result, *domain.Endpoint)
	}
	for _, value := range domain.Endpoints {
		result = append(result, value)
	}
	return result
}
