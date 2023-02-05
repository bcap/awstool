package s3

import (
	awstcmd "awstool/cmd"
	"awstool/cmd/awstool/s3/dump"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
)

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "s3",
		Short:         "Simple Storage Service (S3) related commands",
		SilenceErrors: true,
	}
	awstcmd.AddSubCommand(&cmd, dump.Command(awsCfg))
	return &cmd
}
