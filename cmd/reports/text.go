package reports

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"strings"
)

func ShowTransactions(b *app.BookPrinter, trans []book.Transaction) error {

	// Main transaction (reference one)
	postingTrans := trans[len(trans)-1]

	// First column -- account
	acct_col := make([]string, 0, len(postingTrans))
	acct_col = append(acct_col, b.Ansi(app.BlackUL, "Account"))
	for _, v := range postingTrans {
		lvl := v.GetAccountLevel()
		var t string
		if lvl == 0 {
			t = b.Ansi(app.BlueUL, v.GetAccountTerm())
		} else {
			acctterm := v.GetAccountTerm()
			t = strings.Repeat("  ", lvl) + b.Ansi(app.Black, acctterm)
		}
		acct_col = append(acct_col, t)
	}

	// Second column -- amount
	amt_cols := make([][]string, 0, len(trans))
	for i := range trans {

		idx := make(map[[3]string]int)
		for i, v := range trans[i] {
			idx[[3]string{v.GetAccount(), v.GetAccountTerm(), v.GetCCY()}] = i
		}

		amt_col := make([]string, 0, len(postingTrans))
		amt_col = append(amt_col, b.Ansi(app.BlackUL, trans[i].GetDate().String()))
		for _, p := range postingTrans {
			tidx, ok := idx[[3]string{p.GetAccount(), p.GetAccountTerm(), p.GetCCY()}]
			if !ok {
				return fmt.Errorf("account %s currency %s not on all transactions")
			}

			v := trans[i][tidx]

			if p.GetAccountLevel() != v.GetAccountLevel() || p.GetAccountTerm() != v.GetAccountTerm() {
				return fmt.Errorf("account %s has inconsistent level (%d vs %d) or term (%s vs %s)", p.GetAccount(),
					p.GetAccountLevel(), v.GetAccountLevel(), p.GetAccountTerm(), v.GetAccountTerm())
			}

			amt_col = append(amt_col, b.FormatMoney(v.GetCCY(), v.GetAmount(), 0))
		}
		amt_cols = append(amt_cols, amt_col)
	}

	lacct := app.ListLength(acct_col, 100)

	// Combine amount columns
	namt_col, err := app.Combine(amt_cols, 100)
	if err != nil {
		return err
	}

	// Print final column out
	for i := range acct_col {
		if i > 0 && postingTrans[i-1].GetAccountLevel() == 0 {
			b.Printf("\n")
		}
		b.Printf("%-*.*s %s\n",
			lacct, lacct, acct_col[i],
			namt_col[i])
	}

	return nil
}
