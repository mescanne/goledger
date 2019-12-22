package register

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"strings"
	"unicode/utf8"
)

func ShowRegister(b *app.BookPrinter, inb []book.Transaction, acct string, asc bool) error {
	dates := make([]string, 0, 10)
	payees := make([]string, 0, 10)
	caccts := make([]string, 0, 10)
	amt := make([]string, 0, 10)
	bal := make([]string, 0, 10)

	// Header
	dates = append(dates, b.Ansi(app.UL, "Date"))
	payees = append(payees, b.Ansi(app.UL, "Payee"))
	caccts = append(caccts, b.Ansi(app.UL, "Counteraccount"))
	amt = append(amt, b.Ansi(app.UL, "Amount"))
	bal = append(bal, b.Ansi(app.UL, "Balance"))

	for _, trans := range inb {

		// Find counteraccounts
		accts := make([]string, 0, 1)
		for _, p := range trans {
			if p.GetAccount() == acct {
				continue
			}
			accts = append(accts, p.GetAccount())
		}
		cacct := strings.Join(accts, ",")

		// Find postings for this account (multicurrency possible)
		for _, p := range trans {
			if p.GetAccount() != acct {
				continue
			}
			dates = append(dates, trans.GetDate().String())
			payees = append(payees, trans.GetPayee())
			caccts = append(caccts, cacct)
			amt = append(amt, b.FormatSimpleMoney(p.GetCCY(), p.GetAmount()))
			bal = append(bal, b.FormatSimpleMoney(p.GetCCY(), p.GetBalance()))
		}
	}

	ldate := app.ListLength(dates, 100)
	lpayees := app.ListLength(payees, 100)
	lcaccts := app.ListLength(caccts, 100)
	lamt := app.ListLength(amt, 100)
	lbal := app.ListLength(bal, 100)

	for i := range dates {
		idx := i
		if !asc {
			idx = len(dates) - i - 1
		}
		b.Printf("%s %s %s  %s  %s\n",
			app.PadString(dates[idx], ldate, true),
			app.PadString(payees[idx], lpayees, true),
			app.PadString(caccts[idx], lcaccts, true),
			app.PadString(amt[idx], lamt, false),
			app.PadString(bal[idx], lbal, false))

	}

	return nil
}

func maxLength(strs []string, maxlen int) int {
	l := 0
	for _, s := range strs {
		ls := utf8.RuneCountInString(s)
		if ls > l {
			l = ls
			if l > maxlen {
				return maxlen
			}
		}
	}
	return l
}
