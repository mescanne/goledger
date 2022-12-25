package reports

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"math/big"
	"strings"
)

func FormatCurrency(ccy string) string {
	ccy = strings.ReplaceAll(ccy, " ", "")
	ccy = strings.ToUpper(ccy)
	return ccy
}

func FormatAccount(acct string) string {
	acct = strings.Title(acct)
	return acct
}

func ShowBeancount(b *app.BookPrinter, trans []book.Transaction) error {

	// Gather account info
	type accountInfo struct {
		first  book.Date
		last   book.Date
		isZero bool
	}
	lstats := make(map[string]*accountInfo)
	var zero big.Rat
	for _, t := range trans {
		for _, p := range t {
			acct := FormatAccount(p.GetAccount())
			stats, ok := lstats[acct]
			if !ok {
				stats = &accountInfo{
					first: p.GetDate(),
					last:  p.GetDate(),
				}
				lstats[acct] = stats
			} else {
				stats.last = p.GetDate()
			}
			stats.isZero = (p.GetBalance().Cmp(&zero) == 0)
		}
	}

	for acct, stats := range lstats {
		b.Printf("%s open %s\n", stats.first, acct)
		if stats.isZero {
			b.Printf("%s close %s\n", stats.last.AddDays(1), acct)
		}
	}

	// Dump Account Info for
	// 2009/01/01 open Assets:Mark:Current:CitibankUKChequing
	// 2009/01/01 open Equity:Mark:OpeningBalance
	// DATE close <account>

	// Check all currencies adhere to the rules of beancount, and suggest re-mapping
	// explicitly if not.

	// Transactions
	for _, posts := range trans {
		tnote := posts.GetTransactionNote()
		if tnote != "" {
			tnote = "  ; " + tnote
		}

		// Slightly different format. NOTE -- CCYs have strict requirements.
		b.Printf("%s * \"%s\"%s\n", posts.GetDate(), posts.GetPayee(), tnote)
		for _, p := range posts {
			pnote := p.GetPostNote()
			if pnote != "" {
				pnote = "  ; " + pnote
			}
			ccy := FormatCurrency(p.GetCCY())
			b.Printf("  %s  %s %s%s\n", FormatAccount(p.GetAccount()), b.FormatNumber(p.GetCCY(), p.GetAmount()), ccy, pnote)
		}
		b.Printf("\n")
	}
	return nil
}
