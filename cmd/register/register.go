package register

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"regexp"
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
	Type      string
	Combined  bool
	Macros    []string
	Accounts  []string
	Split     bool
}

const register_long = `Register account postings

Show a registry of postings individual accounts. This is useful for reconciliation
between accounts and for investigating postings.
`

func Add(cmd *cobra.Command, app *app.App, reg *RegisterReport) {
	ncmd := &cobra.Command{
		Args: cobra.MinimumNArgs(1),
		// TODO: Include ops in here as an option
		ValidArgs:         reg.Accounts,
		Use:               "register [macros|ops...] <acct|regex>",
		Long:              register_long,
		Short:             "Show registry of account postings",
		DisableAutoGenTag: true,
	}

	// Set defaults
	reportType := utils.NewEnum(&reg.Type, reportTypes, "reportType")
	ncmd.Flags().Var(reportType, "type", fmt.Sprintf("report type (%s)", reportType.Values()))
	ncmd.Flags().BoolVar(&reg.Combined, "combined", reg.Combined, "combined report (all accounts combined)")
	ncmd.Flags().StringVar(&reg.BeginDate, "begin", reg.BeginDate, "begin date")
	ncmd.Flags().StringVar(&reg.EndDate, "asof", reg.EndDate, "end date")
	ncmd.Flags().IntVar(&reg.Count, "count", reg.Count, "count of entries (0 = no limit)")
	ncmd.Flags().BoolVar(&reg.Asc, "asc", reg.Asc, "ascending or descending order")
	ncmd.Flags().BoolVar(&reg.Split, "split", reg.Split, "split multiple counteraccounts into separate postings")
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

	// Filter by time
	b.FilterByDateSince(book.DateFromString(reg.BeginDate))
	b.FilterByDateAsof(book.DateFromString(reg.EndDate))

	// Apply any operations
	if len(args) > 1 {
		if err = rapp.BookOps(b, args[0:len(args)-1]...); err != nil {
			return err
		}
	}

	// Create printer
	bp := rapp.NewBookPrinter(b.GetCCYDecimals())

	// Combined -- just dump out as is
	arg := args[len(args)-1]
	if reg.Combined {
		re, err := regexp.Compile(arg)
		if err != nil {
			return fmt.Errorf("invalid regex: '%s': %w", arg, err)
		}
		return extractRegisterByRegex(b.Transactions(), re, reg.Split).ShowReport(bp, reg.Type, reg.Count, reg.Asc, true, false)
	}

	for _, acct := range b.Accounts(arg, !rapp.All) {
		bp.Printf("\n%s\n", bp.Ansi(app.BlueUL, acct))
		if err := extractRegisterByAccount(b.Transactions(), acct, reg.Split).ShowReport(bp, reg.Type, reg.Count, reg.Asc, false, true); err != nil {
			return fmt.Errorf("error writing report '%s': %w", acct, err)
		}
	}

	return nil

}
