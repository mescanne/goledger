package currencies

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/spf13/cobra"
)

func Add(root *cobra.Command, app *app.App) {
	ncmd := &cobra.Command{
		Use:               "ccy",
		Short:             "Show currency configuration (decimals)",
		Long:              "Show currency configuration (decimals)",
		DisableAutoGenTag: true,
	}
	ncmd.Args = cobra.NoArgs
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		b, err := app.LoadBook()
		if err != nil {
			return err
		}

		for ccy, dec := range b.GetCCYDecimals() {
			fmt.Fprintf(cmd.OutOrStdout(), "%s => %d decimals\n", ccy, dec)
		}

		return nil
	}
	root.AddCommand(ncmd)
}
