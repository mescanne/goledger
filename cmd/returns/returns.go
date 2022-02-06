package returns

import (
	_ "fmt"
	_ "github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	_ "github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	_ "math/big"
	_ "regexp"
)

// Configuration for a Register Report
type ReturnsReport struct {
	Name  string
	Short string
	Long  string

	TrackAccounts string
	//
	// BeginDate string
	// EndDate   string
	// Count     int
	// Asc       bool
	// Type      string
	// Combined  bool
	// ZeroStart bool
	// Convert   bool
	// Macros    []string
	// Accounts  []string
	// Split     bool
}

const returns_long = `Return account postings

`

func Add(cmd *cobra.Command, app *app.App, ret *ReturnsReport) {
	ncmd := &cobra.Command{
		Args: cobra.MinimumNArgs(1),
		//// TODO: Include ops in here as an option
		// ValidArgs:         reg.Accounts,
		Use:               "returns [macros|ops...] <acct|regex>",
		Long:              returns_long,
		Short:             "Show returns of accounts",
		DisableAutoGenTag: true,
	}

	// Set defaults
	ncmd.Flags().StringVar(&ret.TrackAccounts, "track", ret.TrackAccounts, "tracking accounts")
	//reportType := utils.NewEnum(&reg.Type, reportTypes, "reportType")
	//ncmd.Flags().Var(reportType, "type", fmt.Sprintf("report type (%s)", reportType.Values()))
	//ncmd.Flags().BoolVar(&reg.Combined, "combined", reg.Combined, "combined report (all accounts combined)")
	//ncmd.Flags().StringVar(&reg.BeginDate, "begin", reg.BeginDate, "begin date")
	//ncmd.Flags().StringVar(&reg.EndDate, "asof", reg.EndDate, "end date")
	//ncmd.Flags().IntVar(&reg.Count, "count", reg.Count, "count of entries (0 = no limit)")
	//ncmd.Flags().BoolVar(&reg.Asc, "asc", reg.Asc, "ascending or descending order")
	//ncmd.Flags().BoolVar(&reg.ZeroStart, "zero", reg.ZeroStart, "start balance at zero")
	//ncmd.Flags().BoolVar(&reg.Split, "split", reg.Split, "split multiple counteraccounts into separate postings")
	//ncmd.Flags().BoolVar(&reg.Convert, "convert", reg.Convert, "convert postings to base currency")
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return ret.run(app, cmd, args)
	}

	cmd.AddCommand(ncmd)
}

func (ret *ReturnsReport) run(rapp *app.App, cmd *cobra.Command, args []string) error {

	// Load up saved flags
	b, err := rapp.LoadBook()
	if err != nil {
		return err
	}

	//// Apply any operations
	//if len(args) > 1 {
	//	if err = rapp.BookOps(b, args[0:len(args)-1]...); err != nil {
	//		return err
	//	}
	//}

	// Create printer
	bp := rapp.NewBookPrinter(b.GetCCYDecimals())

	bp.Printf("Hello\n")

	//// Combined -- just dump out as is
	re, err := regexp.Compile(args[0])
	if err != nil {
		return fmt.Errorf("invalid regex: '%s': %w", arg, err)
	}
	tracking, err := regexp.Compile(ret.TrackAccounts)
	if err != nil {
		return fmt.Errorf("invalid tracking regex: '%s': %w", arg, err)
	}

	// Extract it.
	rep := b.ExtractRegister(rapp.BaseCCY, re, true)
	nrep := rep.ExtractSummaryReport(tracking)
	//rep.FilterByDate(reg.BeginDate, reg.EndDate)
	//return ShowReport(bp, rep, reg.Type, reg.Count, reg.Asc, true, true)

	//for _, acct := range b.Accounts(arg, !rapp.All) {
	//	bp.Printf("\n%s\n", bp.Ansi(app.BlueUL, acct))
	//	acctRe, err := regexp.Compile(fmt.Sprintf("^%s$", acct))
	//	if err != nil {
	//		return fmt.Errorf("failed compiling re for account '%s': %w", acct, err)
	//	}
	//	rep = b.ExtractRegister(rapp.BaseCCY, acctRe, reg.Split)
	//	rep.FilterByDate(reg.BeginDate, reg.EndDate)
	//	if err := ShowReport(bp, rep, reg.Type, reg.Count, reg.Asc, false, true); err != nil {
	//		return fmt.Errorf("error writing report '%s': %w", acct, err)
	//	}
	//}

	// Number of columns
	cols := 4

	// ColumnFormats
	fmts := make([]bool, 0, cols)
	fmts = append(fmts, false)
	fmts = append(fmts, true)
	if withAcct {
		fmts = append(fmts, true)
	}
	fmts = append(fmts, true)
	fmts = append(fmts, false)
	if withBal {
		fmts = append(fmts, false)
	}

	// Rows
	rows := make([][]app.ColumnValue, 0, len(report)+1)

	// Header
	header := make([]app.ColumnValue, 0, cols)
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Date")))
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Payee")))
	if withAcct {
		header = append(header, app.ColumnString(b.Ansi(app.UL, "Account")))
	}
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Counteraccount")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Amount")))
	if withBal {
		header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Balance")))
	}
	rows = append(rows, header)

	// Data
	for _, p := range report {
		row := make([]app.ColumnValue, 0, cols)
		row = append(row, app.ColumnString(p.Date.String()))
		row = append(row, app.ColumnString(p.Payee))
		if withAcct {
			row = append(row, app.ColumnString(p.Account))
		}
		row = append(row, app.ColumnString(p.CounterAccount))
		row = append(row, b.GetColumnMoney(p.CCY, p.Amount))
		if withBal {
			row = append(row, b.GetColumnMoney(p.CCY, p.Balance))
		}
		rows = append(rows, row)
	}

	b.PrintColumns(rows, fmts)

	return nil

	return nil

}

//func ShowReport(b *app.BookPrinter, report book.RegistryReport, format string, count int, asc bool, withAcct bool, withBal bool) error {
//
//	// Restrict count counting from beginning
//	if count > 0 && len(report) > count {
//		report = (report)[0:count]
//	}
//
//	// Restrict count counting from end
//	if count < 0 && len(report) > (-1*count) {
//		report = (report)[len(report)+count : len(report)]
//	}
//
//	// Reverse if requested
//	if !asc {
//		ndata := make([]*book.RegistryEntry, len(report), len(report))
//		for i := 0; i < len(report)/2; i++ {
//			ndata[len(report)-i-1] = (report)[i]
//		}
//		report = ndata
//	}
//
//	return ShowText(b, report, withAcct, withBal)
//}
//
//func ShowText(b *app.BookPrinter, report book.RegistryReport, withAcct bool, withBal bool) error {
//
//	// Number of columns
//	cols := 4
//	if withBal {
//		cols++
//	}
//	if withAcct {
//		cols++
//	}
//
//	// ColumnFormats
//	fmts := make([]bool, 0, cols)
//	fmts = append(fmts, false)
//	fmts = append(fmts, true)
//	if withAcct {
//		fmts = append(fmts, true)
//	}
//	fmts = append(fmts, true)
//	fmts = append(fmts, false)
//	if withBal {
//		fmts = append(fmts, false)
//	}
//
//	// Rows
//	rows := make([][]app.ColumnValue, 0, len(report)+1)
//
//	// Header
//	header := make([]app.ColumnValue, 0, cols)
//	header = append(header, app.ColumnString(b.Ansi(app.UL, "Date")))
//	header = append(header, app.ColumnString(b.Ansi(app.UL, "Payee")))
//	if withAcct {
//		header = append(header, app.ColumnString(b.Ansi(app.UL, "Account")))
//	}
//	header = append(header, app.ColumnString(b.Ansi(app.UL, "Counteraccount")))
//	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Amount")))
//	if withBal {
//		header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Balance")))
//	}
//	rows = append(rows, header)
//
//	// Data
//	for _, p := range report {
//		row := make([]app.ColumnValue, 0, cols)
//		row = append(row, app.ColumnString(p.Date.String()))
//		row = append(row, app.ColumnString(p.Payee))
//		if withAcct {
//			row = append(row, app.ColumnString(p.Account))
//		}
//		row = append(row, app.ColumnString(p.CounterAccount))
//		row = append(row, b.GetColumnMoney(p.CCY, p.Amount))
//		if withBal {
//			row = append(row, b.GetColumnMoney(p.CCY, p.Balance))
//		}
//		rows = append(rows, row)
//	}
//
//	b.PrintColumns(rows, fmts)
//
//	return nil
//}
