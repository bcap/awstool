package loader

import (
	"fmt"
)

type FetchBatchFn = func(nextToken *string) (*string, error)

func FetchAll(what string, fetchBatchFn FetchBatchFn) error {
	var nextToken *string
	for batchNum := 1; ; batchNum++ {
		var err error
		var newNextToken *string
		newNextToken, err = fetchBatchFn(nextToken)
		if err != nil {
			return fmt.Errorf(
				"fetching %s in batch %d: %w",
				what, batchNum, err,
			)
		}
		if newNextToken == nil {
			// no more pagination, all fetched
			return nil
		}
		nextToken = newNextToken
	}
}
