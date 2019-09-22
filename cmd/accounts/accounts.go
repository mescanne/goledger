package accounts

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/spf13/cobra"
)

func Add(cmd *cobra.Command, app *app.App) {
	ncmd := &cobra.Command{
		Use:               "accts [regex]",
		Aliases:           []string{"accounts"},
		Short:             "Show matching accounts",
		Long:              "Show matching accounts",
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(ncmd)
	ncmd.Args = cobra.MaximumNArgs(1)
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		b, err := app.LoadBook()
		if err != nil {
			return err
		}
		regex := "^.*$"
		if len(args) == 1 {
			regex = args[0]
		}
		for _, acct := range b.Accounts(regex) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", acct)
		}
		return nil
	}
}
