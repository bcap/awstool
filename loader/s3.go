package loader

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func FetchAllBuckets(ctx context.Context, cfg aws.Config) ([]s3Types.Bucket, error) {
	log.Printf("Fetching all %s S3 buckets", cfg.Region)
	client := s3.NewFromConfig(cfg)
	describeResult, err := client.ListBuckets(
		ctx,
		&s3.ListBucketsInput{},
	)
	if err != nil {
		return nil, err
	}
	log.Printf("Fetched %d %s S3 buckets", len(describeResult.Buckets), cfg.Region)
	return describeResult.Buckets, nil
}

func FetchBucketTags(ctx context.Context, cfg aws.Config, bucket string) ([]s3Types.Tag, error) {
	log.Printf("Fetching tags for %s S3 bucket %s", cfg.Region, bucket)
	client := s3.NewFromConfig(cfg)
	result, err := client.GetBucketTagging(
		ctx,
		&s3.GetBucketTaggingInput{Bucket: &bucket},
	)
	if err != nil {
		return nil, err
	}
	return result.TagSet, nil
}
