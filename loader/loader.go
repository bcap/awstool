package loader

import (
	"fmt"
	"log"
	"math"
	"time"
)

type FetchBatchFn = func(nextToken *string) (*string, error)

func FetchAll(what string, fetchBatchFn FetchBatchFn, options ...LoaderOptionFn) error {
	opts := newLoaderOptions(options...)
	var nextToken *string
	for batchNum := 1; ; batchNum++ {
		var err error
		var newNextToken *string
		for try := 0; ; try++ {
			newNextToken, err = fetchBatchFn(nextToken)
			if err == nil {
				nextToken = newNextToken
				break
			}
			if try == opts.maxRetries || !isRetryable(err) {
				break
			}
			sleepTime := backoffTime(try, opts)
			log.Printf(
				"Got error while fetching %s in batch %d, retrying in %v: %v",
				what, batchNum, sleepTime, err,
			)
			time.Sleep(sleepTime)
		}
		if err != nil {
			return fmt.Errorf(
				"fetching %s in batch %d after %d tries: %w",
				what, batchNum, opts.maxRetries+1, err,
			)
		}
		if nextToken == nil {
			// no more pagination, all fetched
			return nil
		}
	}
}

func backoffTime(try int, opts loaderOptions) time.Duration {
	factor := math.Pow(float64(opts.backoffFactor), float64(try))
	return time.Duration(float64(opts.sleepBetweenRetries) * factor)
}

func isRetryable(err error) bool {
	// TODO differentiate between retryable vs non-retryable errors
	return true
}
