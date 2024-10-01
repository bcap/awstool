package dump

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	client := s3.New(s3.Options{Region: region})
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}
	if downloadVersions {
		return downloadVersioned(ctx, client)
	} else {
		return downloadLatest(ctx, client)
	}
}

func downloadVersioned(ctx context.Context, client *s3.Client) error {
	return nil
}

func downloadLatest(ctx context.Context, client *s3.Client) error {
	return nil
}
