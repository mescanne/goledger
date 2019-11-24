package reports

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"strings"
)

func ShowTransactions(b *app.BookPrinter, trans []book.Transaction) error {

	for _, posts := range trans {

		// Show header
		payee := posts.GetPayee()
		if payee != "" {
			b.Printf("\n%s - %s\n", posts.GetDate(), posts.GetPayee())
		} else {
			b.Printf("\n%s\n", posts.GetDate())
		}

		// First column -- payee
		acct_col := make([]string, 0, 10)
		for _, v := range posts {
			lvl := v.GetAccountLevel()
			var t string
			if lvl == 0 {
				t = b.Ansi(app.BlueUL, v.GetAccountTerm())
			} else {
				t = strings.Repeat("  ", lvl) + b.Ansi(app.Black, v.GetAccountTerm())
			}
			acct_col = append(acct_col, t)
		}

		// Second column -- amount
		amt_col := make([]string, 0, 10)
		for _, v := range posts {
			amt_col = append(amt_col, b.FormatMoney(v.GetCCY(), v.GetAmount(), 0))
		}

		// Format columns
		lacct := app.ListLength(acct_col, 100)
		lamt := app.ListLength(amt_col, 100)
		for i := range acct_col {
			if acct_col[i][0] != ' ' {
				b.Printf("\n")
			}
			b.Printf("%-*.*s %*.*s\n",
				lacct, lacct, acct_col[i],
				lamt, lamt, amt_col[i])
		}
	}

	return nil
}
