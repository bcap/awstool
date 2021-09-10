package ec2

import (
	awstcmd "awstool/cmd"
	"awstool/cmd/awstool/ec2/resolve"
	"awstool/cmd/awstool/ec2/ssh"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
)

func Command(awsCfg **aws.Config) *cobra.Command {
	cmd := cobra.Command{
		Use:           "ec2",
		Short:         "Elastic Compute Cloud related subcommands",
		SilenceErrors: true,
	}
	awstcmd.AddSubCommand(&cmd, resolve.Command(awsCfg))
	awstcmd.AddSubCommand(&cmd, ssh.Command(awsCfg))
	return &cmd
}
