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

const register_long = `Register account postings

Show a registry of postings for an individual account. This
is useful for reconciliation between accounts and for investigating
one account.
`

func Add(cmd *cobra.Command, app *app.App, reg *RegisterReport) {
	ncmd := &cobra.Command{
		Args:              cobra.MinimumNArgs(1),
		ValidArgs:         reg.Accounts,
		Use:               "register [acct regex]...",
		Long:              register_long,
		Short:             "Show registry of account postings",
		DisableAutoGenTag: true,
	}

	// Set defaults
	ncmd.Flags().StringVar(&reg.BeginDate, "begin", reg.BeginDate, "begin date")
	ncmd.Flags().StringVar(&reg.EndDate, "asof", reg.EndDate, "end date")
	ncmd.Flags().IntVar(&reg.Count, "count", reg.Count, "count of entries (0 = no limit)")
	ncmd.Flags().BoolVar(&reg.Asc, "asc", reg.Asc, "ascending or descending order")
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return reg.run(app, cmd, args)
	}

	cmd.AddCommand(ncmd)
}

func (reg *RegisterReport) run(rapp *app.App, cmd *cobra.Command, args []string) error {

	// Load up saved flags
	b, err := rapp.LoadBook()
	if err != nil {
		return err
	}

	b.FilterByDateSince(book.DateFromString(reg.BeginDate))
	b.FilterByDateAsof(book.DateFromString(reg.EndDate))

	bp := rapp.NewBookPrinter(cmd.OutOrStdout(), b.GetCCYDecimals())

	// Show matching accounts
	matched := false
	for _, arg := range args {
		for _, acct := range b.Accounts(arg, !rapp.All) {
			matched = true
			if err := reg.showAccount(b, bp, acct); err != nil {
				return err
			}
		}
	}

	// If nothing matched and we didn't request all.. try again matching
	// everything this time.
	if !matched && !rapp.All {
		for _, arg := range args {
			for _, acct := range b.Accounts(arg, false) {
				if err := reg.showAccount(b, bp, acct); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (reg *RegisterReport) showAccount(b *book.Book, bp *app.BookPrinter, acct string) error {
	regbook := b.Duplicate()
	regbook.FilterByAccount(acct)
	trans := regbook.Transactions()

	// Filter the number of transactions
	if reg.Count > 0 {
		trans = trans[0:reg.Count]
	} else if len(trans)+reg.Count > 0 {
		trans = trans[len(trans)+reg.Count : len(trans)]
	}

	bp.Printf("\n%s\n", bp.Ansi(app.BlueUL, acct))
	if err := ShowRegister(bp, trans, acct, reg.Asc); err != nil {
		return err
	}

	return nil
}
