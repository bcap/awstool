package resolve

import (
	"context"
	"fmt"
	"net/url"
	"sort"
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
	header    bool
	tags      []string
	allTags   bool
	urlEncode bool
	template  *template.Template
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
	var noURLEncode bool

	cmd.Flags().StringVarP(
		&domain, "domain", "d", "",
		"Find the domain by its name",
	)

	cmd.Flags().StringSliceVarP(
		&printOptions.tags, "print-tags", "T", []string{},
		"Pass a list of tags keys that should be printed. See also --print-all-tags",
	)

	cmd.Flags().BoolVarP(
		&printOptions.allTags, "print-all-tags", "A", false,
		"Also prints all tags associated to the domains. Overrides --print-tags",
	)

	cmd.Flags().BoolVarP(
		&printOptions.header, "header", "H", false,
		"Also print a header on the first line, which will name the columns being printed",
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
		printOptions.urlEncode = !noURLEncode
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

type printDomainInput struct {
	region string
	domain *awst.ElasticsearchDomain
}

func printDomains(aws *awst.AWS, printOptions printOptions) {
	printHeader(printOptions)

	inputs := []printDomainInput{}
	for _, region := range aws.Regions {
		for _, domain := range region.Elasticsearch.Domains {
			inputs = append(inputs, printDomainInput{
				region: region.Region,
				domain: domain,
			})
		}
	}

	sort.SliceStable(inputs, func(i, j int) bool {
		comparison := strings.Compare(inputs[i].region, inputs[j].region)
		if comparison == 0 {
			comparison = strings.Compare(
				*inputs[i].domain.Status.DomainName,
				*inputs[j].domain.Status.DomainName,
			)
		}
		return comparison < 0
	})

	for _, input := range inputs {
		printDomain(input.region, input.domain, printOptions)
	}
}

func printHeader(printOptions printOptions) {
	if !printOptions.header || printOptions.template != nil {
		return
	}
	fmt.Print("#region #domain #endpoints #instance_count #instance_type")
	if printOptions.allTags || len(printOptions.tags) > 0 {
		fmt.Print(" #tags")
	}
	fmt.Println()
}

type templateData struct {
	Region string
	Domain *awst.ElasticsearchDomain
}

func printDomain(region string, domain *awst.ElasticsearchDomain, printOptions printOptions) {
	if printOptions.template != nil {
		buf := strings.Builder{}
		data := templateData{
			Region: region,
			Domain: domain,
		}
		if err := printOptions.template.Execute(&buf, data); err != nil {
			fmt.Printf("template execution error: %v\n", err)
		}
		fmt.Println(buf.String())
		return
	}

	fmt.Printf(
		"%s %s %s %d %s",
		region,
		*domain.Status.DomainName,
		strings.Join(endpoints(domain.Status), ","),
		*domain.Status.ElasticsearchClusterConfig.InstanceCount,
		domain.Status.ElasticsearchClusterConfig.InstanceType,
	)
	tagsStr := tagsString(domain, printOptions)
	if tagsStr != "" {
		fmt.Printf(" %s", tagsStr)
	}
	fmt.Println()
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

func tagsString(domain *awst.ElasticsearchDomain, printOptions printOptions) string {
	if !printOptions.allTags && len(printOptions.tags) == 0 {
		return ""
	}

	result := []string{}
	appendTag := func(tag esTypes.Tag) {
		value := *tag.Value
		if printOptions.urlEncode {
			value = url.PathEscape(value)
		}
		result = append(result, *tag.Key+":"+value)
	}

	if printOptions.allTags {
		for _, tag := range domain.Tags {
			appendTag(tag)
		}

	} else {
		tagMap := map[string]esTypes.Tag{}
		for _, tag := range domain.Tags {
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
