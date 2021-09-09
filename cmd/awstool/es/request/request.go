package request

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	awst "awstool/aws"
	"awstool/aws/elasticsearch"
	"awstool/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Command(awsCfg **aws.Config) *cobra.Command {
	expectedPositionals := "REGION DOMAIN METHOD PATH [DATA]"

	cmd := cobra.Command{
		Use:   "request",
		Short: "submits a request to given elasticsearch domain",
		Long: "Submits a request to an elasticsearch domain by the given inputs. " +
			"Arguments must go in following order: " + expectedPositionals,
		SilenceErrors: true,
	}

	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) < 4 || len(args) > 5 {
			return fmt.Errorf(
				"incorrect number of args passed. "+
					"Positional args are expected to be in the following format: %s",
				expectedPositionals,
			)
		}
		return nil
	}

	var headers []string

	cmd.Flags().StringSliceVarP(
		&headers, "headers", "H", []string{},
		"Headers to be passed in the request. Format is key:value for each header. "+
			"Can be specified multiple times or a single time with comma separated values",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		region := args[0]
		domainName := args[1]
		method := args[2]
		path := args[3]

		var data []byte
		if len(args) > 4 {
			var err error
			data, err = loadDataArg(args[4])
			if err != nil {
				return err
			}
		}

		headerMap, err := toHeaderMap(headers)
		if err != nil {
			return err
		}

		// We silence usage here instead of setting in the command struct declaration because it is
		// only at this point forward that we want to not display the usage when an error occurs,
		// as it will be an execution error, not a parsing/usage error
		// See more at https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		domain, err := resolve(cmd.Context(), **awsCfg, region, domainName)
		if err != nil {
			return err
		}

		_, err = request(cmd.Context(), domain, method, path, headerMap, data)
		return err
	}

	return &cmd
}

func resolve(ctx context.Context, cfg aws.Config, region string, domain string) (*awst.ElasticsearchDomain, error) {
	resolution, err := loader.LoadAWS(
		ctx, cfg,
		loader.WithServices("elasticsearch"),
		loader.WithRegions(region),
		loader.WithESFetchOptions(elasticsearch.WithDomains(domain)),
	)
	if err != nil {
		return nil, err
	}
	var domains []*awst.ElasticsearchDomain
	for _, region := range resolution.Regions {
		for _, domain := range region.Elasticsearch.Domains {
			domains = append(domains, domain)
		}
	}
	if len(domains) == 0 {
		return nil, fmt.Errorf("no domain found")
	}
	if len(domains) > 1 {
		domainNames := make([]string, len(domains))
		for idx, domain := range domains {
			domainNames[idx] = *domain.Status.DomainName
		}
		return nil, fmt.Errorf("multiple domains found: %v", domainNames)
	}
	return domains[0], nil
}

func request(ctx context.Context, domain *awst.ElasticsearchDomain, method string, path string, headers map[string]string, data []byte) (*http.Response, error) {
	client := http.DefaultClient
	endpoint := domain.Status.Endpoint
	if endpoint == nil {
		for _, v := range domain.Status.Endpoints {
			endpoint = &v
		}
	}
	if endpoint == nil {
		return nil, fmt.Errorf("could not resolve endpoint for domain")
	}

	path = sanitizePath(path)
	urlString := "https://" + *endpoint + "/" + path
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("invalid generated url: %s: %w", urlString, err)
	}

	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	printRequest(req)

	start := time.Now()
	resp, err := client.Do(req)

	printResponse(resp, start)

	return resp, err
}

func printRequest(req *http.Request) {
	if !log.IsLevelEnabled(log.InfoLevel) {
		return
	}
	fmt.Fprintf(os.Stderr, "> %s %s\n", req.Method, req.URL.Path)
	fmt.Fprintf(os.Stderr, "> Host: %s\n", req.URL.Host)
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(os.Stderr, "> %s: %s\n", key, value)
		}
	}
	fmt.Fprint(os.Stderr, ">\n")
}

func printResponse(resp *http.Response, reqTime time.Time) error {
	if log.IsLevelEnabled(log.InfoLevel) {
		fmt.Fprintf(os.Stderr, "< %s\n", resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(os.Stderr, "< %s: %s\n", key, value)
			}
		}
		fmt.Fprint(os.Stderr, "<\n")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if log.IsLevelEnabled(log.InfoLevel) {
		fmt.Fprintf(os.Stderr, "* response received in %v\n", time.Since(reqTime))
	}

	fmt.Print(string(body))

	return nil
}

func loadDataArg(arg string) ([]byte, error) {
	if arg[0] == '@' {
		filename := arg[1:]
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", filename, err)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		return data, nil
	}
	return []byte(arg), nil
}

func toHeaderMap(headers []string) (map[string]string, error) {
	result := map[string]string{}
	for _, header := range headers {
		colonIdx := strings.Index(header, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid header %q: no colon separating values", header)
		}
		key := strings.TrimSpace(header[:colonIdx])
		value := strings.TrimSpace(header[colonIdx+1:])
		if key == "" {
			return nil, fmt.Errorf("invalid header %q: key is empty", header)
		}
		if value == "" {
			return nil, fmt.Errorf("invalid header %q: value is empty", header)
		}
		result[key] = value
	}
	return result, nil
}

func sanitizePath(path string) string {
	for {
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		} else {
			return path
		}
	}
}
