package register

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"strings"
)

func ShowRegister(b *app.BookPrinter, inb []book.Transaction, acct string, asc bool) error {
	dates := make([]string, 0, 10)
	payees := make([]string, 0, 10)
	caccts := make([]string, 0, 10)
	amt := make([]string, 0, 10)
	bal := make([]string, 0, 10)

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

	payees = formatStrings(payees, 100, true)
	caccts = formatStrings(caccts, 100, true)
	amt = formatStrings(amt, 100, false)
	bal = formatStrings(bal, 100, false)

	return writeStrings(b, asc, dates, payees, caccts, amt, bal)
}

func writeStrings(b *app.BookPrinter, asc bool, icols ...[]string) error {

	if len(icols) == 0 {
		return fmt.Errorf("invalid - no columns to write")
	}

	cols := len(icols)
	rows := len(icols[0])

	for _, c := range icols[1:] {
		if len(c) != rows {
			return fmt.Errorf("invalid - columns of different row size - %d vs %d", len(c), rows)
		}
	}

	i := 0
	if !asc {
		i = rows - 1
	}
	for {
		for c, col := range icols {
			b.Printf("%s", col[i])
			if c < (cols - 1) {
				b.Printf(" ")
			} else {
				b.Printf("\n")
			}
		}

		if asc {
			i++
			if i >= rows {
				break
			}
		} else {
			i--
			if i < 0 {
				break
			}
		}
	}

	return nil
}

func formatStrings(strs []string, maxLen int, isLeft bool) []string {
	l := 0
	for _, s := range strs {
		ls := len(s)
		if ls > l {
			l = ls
		}
	}

	if l > maxLen && maxLen > 0 {
		l = maxLen
	}

	ostrs := make([]string, len(strs))
	for i, s := range strs {
		if isLeft {
			ostrs[i] = fmt.Sprintf("%-*.*s", l, l, s)
		} else {
			ostrs[i] = fmt.Sprintf("%*.*s", l, l, s)
		}
	}

	return ostrs
}
