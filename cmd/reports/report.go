package reports

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"math/big"
	"regexp"
)

type TransactionReport struct {
	Credit     string
	Hidden     string
	Convert    bool
	JsonPretty bool
	HTMLCSS    string
	Sum        bool
	Type       string
	Combineby  string
}

const (
	custom_func = `
__goledger_handle_noun() {
    c=$((c+1))
}
`
)

const report_long = `Aggregated transactions reports

There are two basic dimensions for transactions reports:

Time period:
  If you look at a time period from the beginning of time
  until a point in time (eg now) you will see the total
  balance of all of the accounts.
  
  If you select a start and end date (eg beginning of this
  year until now) you will see the change in balance
  across all of the accounts.

Account Regexp:
  Normally, you don't want to see all accounts but focus
  on a particular subset of the accounts. Or a certain
  categorisation of accounts.
  
  Using regular expressions you can create income statements,
  balance sheets, and cashflow statements.

  Example for a balance sheet:
  Map all ^Income:.* and ^Expense.:* accounts into Equity. Also
  include all other accounts that aren't Asset:, Liability:,
  or Equity.

  This will leave just Asset and Liabilities.
    
`

func Add(cmd *cobra.Command, app *app.App, report *TransactionReport) {
	ncmd := &cobra.Command{
		Aliases:           []string{"trans", "transactions"},
		Use:               "report [macros|ops...]",
		Short:             "Aggregated transaction reports",
		Long:              report_long,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(ncmd)

	// Set defaults
	floorType := utils.NewEnum(&report.Combineby, book.FloorTypes, "floorType")
	ncmd.Flags().Var(floorType, "splitby", fmt.Sprintf("combine transactions by periodic date (values %s)", floorType.Values()))
	reportType := utils.NewEnum(&report.Type, []string{"Text", "Ledger", "Json", "HTML"}, "reportType")
	ncmd.Flags().Var(reportType, "type", fmt.Sprintf("report type (%s)", reportType.Values()))
	ncmd.Flags().BoolVar(&report.Sum, "sum", report.Sum, "summarise transactions")
	ncmd.Flags().BoolVar(&report.Convert, "convert", report.Convert, "convert to base currency")
	ncmd.Flags().BoolVar(&report.JsonPretty, "jsonpretty", report.JsonPretty, "pretty Json (indented) for Json output")
	ncmd.Flags().StringVar(&report.HTMLCSS, "htmlcss", report.HTMLCSS, "HTML CSS (string or file:<css file>) for HTML output (inlined in HTML)")
	ncmd.Flags().StringVar(&report.Credit, "credit", report.Credit, "credit account regex for summary")
	ncmd.Flags().StringVar(&report.Hidden, "hidden", report.Hidden, "hidden account in reports for summary")

	// don't need to save it
	macroNames := make([]string, 0, len(app.Macros))
	for k, _ := range app.Macros {
		macroNames = append(macroNames, k)
	}
	ncmd.ValidArgs = macroNames
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return report.run(app, cmd, args)
	}
}

func (report *TransactionReport) run(app *app.App, cmd *cobra.Command, args []string) error {

	// Load up saved flags
	b, err := app.LoadBook()
	if err != nil {
		return err
	}

	// Apply ops
	err = app.BookOps(b, args...)
	if err != nil {
		return err
	}

	// Always combine for reports
	b.SplitBy(report.Combineby)

	if report.Convert {
		if app.BaseCCY == "" {
			return fmt.Errorf("unable to convert -- no CCY specified")
		}
		b.MapAmount(func(date book.Date, iccy string) (*big.Rat, string) {
			rate, _ := b.GetPrice(date, iccy, app.BaseCCY)
			return rate, app.BaseCCY
		})
	}

	var creditre *regexp.Regexp = nil
	if report.Credit != "" {
		creditre, err = regexp.Compile(report.Credit)
		if err != nil {
			return fmt.Errorf("failed compiling credit accounts '%s': %w", report.Credit, err)
		}
	}

	var trans []book.Transaction
	if report.Sum {
		if app.BaseCCY == "" {
			return fmt.Errorf("unable to convert -- no CCY specified")
		}
		trans = b.Accumulate(app.BaseCCY, app.Divider, creditre, report.Hidden)
	} else {
		trans = b.Transactions()
	}

	bp := app.NewBookPrinter(b.GetCCYDecimals())

	// Need type of report now..
	if report.Type == "Text" {
		return ShowTransactions(bp, trans)
	} else if report.Type == "Json" {
		return bp.PrintJSON(trans, report.JsonPretty)
	} else if report.Type == "HTML" {
		return ShowHTMLTransactions(bp, trans, report.HTMLCSS)
	} else {
		return ShowLedger(bp, trans)
	}
}
