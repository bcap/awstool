package es

import (
	awstcmd "awstool/cmd"
	"awstool/cmd/awstool/es/request"
	"awstool/cmd/awstool/es/resolve"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
)

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "es",
		Short:         "Elasticsearch related subcommands",
		SilenceErrors: true,
	}
	awstcmd.AddSubCommand(&cmd, resolve.Command(awsCfg))
	awstcmd.AddSubCommand(&cmd, request.Command(awsCfg))
	return &cmd
}
