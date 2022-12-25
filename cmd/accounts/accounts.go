package accounts

import (
	rapp "github.com/mescanne/goledger/cmd/app"
	"github.com/spf13/cobra"
)

func Add(cmd *cobra.Command, app *rapp.App) {
	ncmd := &cobra.Command{
		Use:               "accts [regex]",
		Aliases:           []string{"accounts"},
		Short:             "Show matching accounts",
		Long:              "Show matching accounts",
		DisableAutoGenTag: true,
	}
	ncmd.Args = cobra.MaximumNArgs(1)

	// Set defaults
	var useJson bool
	ncmd.Flags().BoolVar(&useJson, "json", false, "Show accounts using json")

	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		b, err := app.LoadBook()
		if err != nil {
			return err
		}
		regex := "^.*$"
		if len(args) == 1 {
			regex = args[0]
		}

		// Create printer
		bp := app.NewBookPrinter(b.GetCCYDecimals())

		if useJson {
			accts := make([]string, 0, 100)
			for _, acct := range b.Accounts(regex, !app.All) {
				accts = append(accts, acct)
			}
			bp.PrintJSON(accts, true)
		} else {
			rows := make([][]rapp.ColumnValue, 0, 100)
			rows = append(rows, []rapp.ColumnValue{rapp.ColumnString(bp.Ansi(rapp.BlueUL, "Account"))})
			for _, acct := range b.Accounts(regex, !app.All) {
				rows = append(rows, []rapp.ColumnValue{rapp.ColumnString(acct)})
			}
			bp.PrintColumns(rows, []bool{false})
		}

		return nil
	}

	cmd.AddCommand(ncmd)
}
