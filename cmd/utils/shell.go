package utils

import (
	"fmt"
	"github.com/spf13/cobra"
)

var shell_long = `Shell integration

This integrates into the shell. For bash for example:

  eval "$(goledger shell --type=bash)"

This is designed to make it much easier to use goledger
from the command line.
`

func AddShell(root *cobra.Command) {
	ncmd := &cobra.Command{
		Use:               "shell",
		Short:             "Shell integration",
		Aliases:           []string{"completions"},
		Long:              shell_long,
		DisableAutoGenTag: true,
	}

	var shell string
	shells := []string{"bash", "zsh", "powershell"}
	shellType := NewEnum(&shell, shells, "shell")
	ncmd.Flags().Var(shellType, "type", fmt.Sprintf("Shell for integration (values %s)", shellType.Values()))
	ncmd.ValidArgs = shells

	ncmd.PreRunE = cobra.NoArgs
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		if shell == "bash" {
			fmt.Fprintf(rcmd.OutOrStdout(), "COMP_WORDBREAKS=${COMP_WORDBREAKS//:}\n")
			return root.GenBashCompletion(rcmd.OutOrStdout())
		} else if shell == "zsh" {
			fmt.Fprintf(rcmd.OutOrStdout(), "COMP_WORDBREAKS=${COMP_WORDBREAKS//:}\n")
			return root.GenZshCompletion(rcmd.OutOrStdout())
		} else if shell == "powershell" {
			return root.GenPowerShellCompletion(rcmd.OutOrStdout())
		}
		return fmt.Errorf("invalid shell type %s: must be one of %s", shell, shellType.Values())
	}
	root.AddCommand(ncmd)
}
