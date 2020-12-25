package register

import (
	"encoding/csv"
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

	// Initialize columns
	dates := make([]string, 0, 100)
	payees := make([]string, 0, 100)
	accts := make([]string, 0, 100)
	caccts := make([]string, 0, 100)
	amt := make([]string, 0, 100)
	bal := make([]string, 0, 100)

	// Header
	dates = append(dates, b.Ansi(app.UL, "Date"))
	payees = append(payees, b.Ansi(app.UL, "Payee"))
	accts = append(accts, b.Ansi(app.UL, "Account"))
	caccts = append(caccts, b.Ansi(app.UL, "Counteraccount"))
	amt = append(amt, b.Ansi(app.UL, "Amount"))
	bal = append(bal, b.Ansi(app.UL, "Balance"))

	for _, p := range report {
		dates = append(dates, p.Date.String())
		payees = append(payees, p.Payee)
		accts = append(accts, p.Account)
		caccts = append(caccts, strings.Join(p.CounterAccount, ";"))
		amt = append(amt, b.FormatSimpleMoney(p.CCY, p.Amount))
		bal = append(bal, b.FormatSimpleMoney(p.CCY, p.Balance))
	}

	// Apply formatting
	ldate := app.ListLength(dates, 100)
	lpayees := app.ListLength(payees, 100)
	laccts := app.ListLength(accts, 100)
	lcaccts := app.ListLength(caccts, 100)
	lamt := app.ListLength(amt, 100)
	lbal := app.ListLength(bal, 100)

	// Apply padding and print
	for idx := range dates {
		if withAcct {
			b.Printf("%s %s %s %s %s",
				app.PadString(dates[idx], ldate, true),
				app.PadString(accts[idx], laccts, true),
				app.PadString(payees[idx], lpayees, true),
				app.PadString(caccts[idx], lcaccts, true),
				app.PadString(amt[idx], lamt, false))
		} else {
			b.Printf("%s %s %s %s",
				app.PadString(dates[idx], ldate, true),
				app.PadString(payees[idx], lpayees, true),
				app.PadString(caccts[idx], lcaccts, true),
				app.PadString(amt[idx], lamt, false))
		}
		if withBal {
			b.Printf(" %s", app.PadString(bal[idx], lbal, false))
		}
		b.Printf("\n")
	}

	return nil
}

func (report registryReport) ShowCSV(b *app.BookPrinter) error {

	csvwrite := csv.NewWriter(b)
	err := csvwrite.Write([]string{
		"date",
		"payee",
		"account",
		"counterAccount",
		"ccy",
		"amount",
		"balance",
	})
	if err != nil {
		return fmt.Errorf("error writing csv: %w", err)
	}

	// Print out content
	for _, p := range report {
		amt, _ := p.Amount.Float64()
		bal, _ := p.Balance.Float64()
		err := csvwrite.Write([]string{
			p.Date.String(),
			p.Payee,
			p.Account,
			strings.Join(p.CounterAccount, ";"),
			p.CCY,
			fmt.Sprintf("%f", amt),
			fmt.Sprintf("%f", bal),
		})
		if err != nil {
			return fmt.Errorf("error writing csv: %w", err)
		}
	}

	csvwrite.Flush()

	return nil
}
