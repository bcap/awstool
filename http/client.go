package http

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"golang.org/x/sync/semaphore"
)

type ParallelLimitedHTTPClient struct {
	config.HTTPClient
	sem *semaphore.Weighted
}

func NewParallelLimitedHTTPClient(client config.HTTPClient, maxParallelism int) *ParallelLimitedHTTPClient {
	return &ParallelLimitedHTTPClient{
		HTTPClient: client,
		sem:        semaphore.NewWeighted(int64(maxParallelism)),
	}
}

func (p *ParallelLimitedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	p.sem.Acquire(req.Context(), 1)
	defer p.sem.Release(1)
	return p.HTTPClient.Do(req)
}
