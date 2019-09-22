package register

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/spf13/cobra"
)

// Configuration for a Register Report
type RegisterReport struct {
	Name      string
	Short     string
	Long      string
	BeginDate string
	EndDate   string
	Count     int
	Asc       bool
	Macros    []string
	Accounts  []string
}

func Add(cmd *cobra.Command, app *app.App, reg *RegisterReport) {
	ncmd := &cobra.Command{
		Args:              cobra.MinimumNArgs(1),
		ValidArgs:         reg.Accounts,
		Use:               "register [acct regex]...",
		Long:              "long reg\nmultiline\n",
		Short:             "short reg",
		DisableAutoGenTag: true,
	}

	// Set defaults
	ncmd.Flags().StringVar(&reg.BeginDate, "begin", reg.BeginDate, "Begin date")
	ncmd.Flags().StringVar(&reg.EndDate, "asof", reg.EndDate, "End date")
	ncmd.Flags().IntVar(&reg.Count, "count", reg.Count, "Count of entries (0 = no limit)")
	ncmd.Flags().BoolVar(&reg.Asc, "asc", reg.Asc, "Ascending or descending order")
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return reg.run(app, cmd, args)
	}

	cmd.AddCommand(ncmd)
}

func (reg *RegisterReport) run(app *app.App, cmd *cobra.Command, args []string) error {

	// Load up saved flags
	b, err := app.LoadBook()
	if err != nil {
		return err
	}

	b.FilterByDateSince(book.DateFromString(reg.BeginDate))
	b.FilterByDateAsof(book.DateFromString(reg.EndDate))

	bp := app.NewBookPrinter(cmd.OutOrStdout(), b.GetCCYDecimals())

	// Show all matching accounts
	for _, arg := range args {
		for _, acct := range b.Accounts(arg) {

			// Get the translations for this account
			regbook := b.Duplicate()
			regbook.FilterByAccount(acct)
			trans := regbook.Transactions()

			// Filter the number of transactions
			if reg.Count > 0 {
				trans = trans[0:reg.Count]
			} else {
				trans = trans[len(trans)+reg.Count : len(trans)]
			}

			bp.Printf("\n%s\n", bp.BlueUL(acct))
			if err := ShowRegister(bp, trans, acct, reg.Asc); err != nil {
				return err
			}
		}
	}

	return nil
}
