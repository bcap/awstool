package cmd

import (
	"github.com/spf13/cobra"
)

func AddSubCommand(cmd *cobra.Command, subCmd *cobra.Command) {
	if subCmd.PersistentPreRun != nil {
		innerFn := subCmd.PersistentPreRun
		subCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
			if cmd.Parent().PersistentPreRun != nil {
				cmd.Parent().PersistentPreRun(cmd, args)
			}
			innerFn(cmd, args)
		}
	}
	if subCmd.PersistentPreRunE != nil {
		innerFn := subCmd.PersistentPreRunE
		subCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PersistentPreRunE != nil {
				if err := cmd.Parent().PersistentPreRunE(cmd, args); err != nil {
					return err
				}
			}
			return innerFn(cmd, args)
		}
	}
	if subCmd.PersistentPostRun != nil {
		innerFn := subCmd.PersistentPostRun
		subCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
			if cmd.Parent().PersistentPostRun != nil {
				cmd.Parent().PersistentPostRun(cmd, args)
			}
			innerFn(cmd, args)
		}
	}
	if subCmd.PersistentPostRunE != nil {
		innerFn := subCmd.PersistentPostRunE
		subCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PersistentPostRunE != nil {
				if err := cmd.Parent().PersistentPostRunE(cmd, args); err != nil {
					return err
				}
			}
			return innerFn(cmd, args)
		}
	}
	cmd.AddCommand(subCmd)
}
