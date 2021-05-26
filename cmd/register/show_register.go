package register

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"strings"
)

var reportTypes = []string{
	"Text",
	"JSON",
	"CSV",
}

func ShowReport(b *app.BookPrinter, report book.RegistryReport, format string, count int, asc bool, withAcct bool, withBal bool) error {

	// Restrict count counting from beginning
	if count > 0 && len(report) > count {
		report = (report)[0:count]
	}

	// Restrict count counting from end
	if count < 0 && len(report) > (-1*count) {
		report = (report)[len(report)+count : len(report)]
	}

	// Reverse if requested
	if !asc {
		ndata := make([]*book.RegistryEntry, len(report), len(report))
		for i := 0; i < len(report)/2; i++ {
			ndata[len(report)-i-1] = (report)[i]
		}
		report = ndata
	}

	if format == "Text" {
		return ShowText(b, report, withAcct, withBal)
	} else if format == "JSON" {
		return b.PrintJSON(report, true)
	} else if format == "CSV" {
		return ShowCSV(b, report)
	} else {
		return fmt.Errorf("invalid report type '%s', expected %s", format, strings.Join(reportTypes, ", "))
	}
}

func ShowText(b *app.BookPrinter, report book.RegistryReport, withAcct bool, withBal bool) error {

	// Number of columns
	cols := 4
	if withBal {
		cols++
	}
	if withAcct {
		cols++
	}

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
}
func ShowCSV(b *app.BookPrinter, report book.RegistryReport) error {

	rows := make([][]string, 0, len(report)+1)

	rows = append(rows, []string{
		"date",
		"payee",
		"account",
		"counterAccount",
		"ccy",
		"amount",
		"balance",
	})

	// Print out content
	for _, p := range report {
		amt, _ := p.Amount.Float64()
		bal, _ := p.Balance.Float64()
		rows = append(rows, []string{
			p.Date.String(),
			p.Payee,
			p.Account,
			p.CounterAccount,
			p.CCY,
			fmt.Sprintf("%f", amt),
			fmt.Sprintf("%f", bal),
		})
	}

	return b.PrintCSV(rows)
}
