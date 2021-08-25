package s3

import (
	"context"
	"log"

	"aws-tools/loader"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func FetchAllBuckets(
	ctx context.Context,
	cfg aws.Config,
	loaderOpts ...loader.LoaderOptionFn,
) ([]s3Types.Bucket, error) {
	log.Print("Fetching all S3 buckets")

	result := []s3Types.Bucket{}

	client := s3.NewFromConfig(cfg)

	load := func(nextToken *string) (*string, error) {
		describeResult, err := client.ListBuckets(
			ctx,
			&s3.ListBucketsInput{},
		)
		if err != nil {
			return nil, err
		}
		result = append(result, describeResult.Buckets...)
		return nil, nil
	}

	err := loader.FetchAll("S3 buckets", load, loaderOpts...)
	if err != nil {
		return result, err
	}

	log.Printf("Fetched %d S3 buckets", len(result))

	return result, nil
}
