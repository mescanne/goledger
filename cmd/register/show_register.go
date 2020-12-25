package register

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"math/big"
	"regexp"
	"strings"
)

var reportTypes = []string{
	"Text",
	"JSON",
	"CSV",
}

type registryReport []*registryEntry
type registryEntry struct {
	Date           book.Date `json:"date"`
	Account        string    `json:"account"`
	Payee          string    `json:"payee,omitempty"`
	CounterAccount []string  `json:"counterAccount,omitempty"`
	Amount         *big.Rat  `json:"amount"`
	CCY            string    `json:"ccy"`
	Balance        *big.Rat  `json:"balance"`
}

func extractRegisterReport(t book.Transaction, p book.Posting, split bool) []*registryEntry {

	// Algorithm:
	//  - Find all other postings that are of the same currency in the opposite direction
	//  - Sum up those values
	//  - Re-iterate and create new postings in the same portion of value/sum.

	// Find counteraccounts
	// Only counteraccounts of the same currency in the opposite direction are of interest!
	caccts := make([]string, 0, 1)
	camts := make([]*big.Rat, 0, 1)
	sumamt := big.NewRat(0, 1)
	for _, tp := range t {
		if tp.GetAccount() == p.GetAccount() {
			continue
		}
		if tp.GetAmount().Sign() == p.GetAmount().Sign() {
			continue
		}
		if tp.GetCCY() != p.GetCCY() {
			continue
		}
		caccts = append(caccts, tp.GetAccount())
		camts = append(camts, tp.GetAmount())
		sumamt.Add(sumamt, tp.GetAmount())
	}

	// Invert it
	sumamt.Inv(sumamt)

	// Simple case - only one counteraccount, general scenario
	// ... or the flag indicates to do it anyway.
	if !split || len(caccts) == 1 {
		return []*registryEntry{
			&registryEntry{
				Date:           t.GetDate(),
				Payee:          t.GetPayee(),
				Account:        p.GetAccount(),
				CounterAccount: caccts,
				Amount:         p.GetAmount(),
				CCY:            p.GetCCY(),
				Balance:        p.GetBalance(),
			},
		}
	}

	// Allocate them
	entries := make([]*registryEntry, len(caccts), len(caccts))
	for i, v := range caccts {

		// Amount we want -
		// (p.GetAmount() * camts[i]) / sumamt)
		namt := big.NewRat(0, 1)
		namt.Mul(p.GetAmount(), camts[i])
		namt.Mul(namt, sumamt)

		entries[i] = &registryEntry{
			Date:           t.GetDate(),
			Payee:          t.GetPayee(),
			Account:        p.GetAccount(),
			CounterAccount: []string{v},
			Amount:         namt,
			CCY:            p.GetCCY(),
			Balance:        p.GetBalance(),
		}
	}

	return entries
}

func extractRegisterByAccount(inb []book.Transaction, acct string, split bool) registryReport {
	data := make([]*registryEntry, 0, 100)

	// Iterate each transaction
	for _, trans := range inb {

		// Iterate each posting
		for _, p := range trans {

			// If this isn't eligible for printing, skip it
			if p.GetAccount() != acct {
				continue
			}

			data = append(data, extractRegisterReport(trans, p, split)...)
		}
	}

	return data
}

func extractRegisterByRegex(inb []book.Transaction, re *regexp.Regexp, split bool) registryReport {
	data := make([]*registryEntry, 0, 100)

	// Iterate each transaction
	for _, trans := range inb {

		// Iterate each posting
		for _, p := range trans {

			// If this isn't eligible for printing, skip it
			if !re.MatchString(p.GetAccount()) {
				continue
			}

			data = append(data, extractRegisterReport(trans, p, split)...)
		}
	}

	return data
}

func (report registryReport) ShowReport(b *app.BookPrinter, format string, count int, asc bool, withAcct bool, withBal bool) error {

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
		ndata := make([]*registryEntry, len(report), len(report))
		for i := 0; i < len(report)/2; i++ {
			ndata[len(report)-i-1] = (report)[i]
		}
		report = ndata
	}

	if format == "Text" {
		return report.ShowText(b, withAcct, withBal)
	} else if format == "JSON" {
		return b.PrintJSON(report, true)
	} else if format == "CSV" {
		return report.ShowCSV(b)
	} else {
		return fmt.Errorf("invalid report type '%s', expected %s", format, strings.Join(reportTypes, ", "))
	}
}

func (report registryReport) ShowText(b *app.BookPrinter, withAcct bool, withBal bool) error {

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
		row = append(row, app.ColumnString(strings.Join(p.CounterAccount, ";")))
		row = append(row, b.GetColumnMoney(p.CCY, p.Amount))
		if withBal {
			row = append(row, b.GetColumnMoney(p.CCY, p.Balance))
		}
		rows = append(rows, row)
	}

	b.PrintColumns(rows, fmts)

	return nil
}
func (report registryReport) ShowCSV(b *app.BookPrinter) error {

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
			strings.Join(p.CounterAccount, ";"),
			p.CCY,
			fmt.Sprintf("%f", amt),
			fmt.Sprintf("%f", bal),
		})
	}

	return b.PrintCSV(rows)
}
