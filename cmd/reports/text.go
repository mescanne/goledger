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
	acct_col := make([]string, len(postingTrans)+1)
	acct_col[0] = b.Ansi(app.UL, "Account")
	for i, v := range postingTrans {
		lvl := v.GetAccountLevel()
		var t string
		if lvl == 0 {
			t = b.Ansi(app.BlueUL, v.GetAccountTerm())
		} else {
			acctterm := v.GetAccountTerm()
			t = strings.Repeat("  ", lvl) + acctterm
		}
		acct_col[i+1] = t
	}

	// Maximum length for amounts
	maxlen := 0
	for _, t := range trans {
		for _, p := range t {
			l := app.Length(b.FormatSimpleMoney(p.GetCCY(), p.GetAmount()))
			if l > maxlen {
				maxlen = l
			}
		}
	}

	header := make([]string, len(trans))
	for i, t := range trans {
		header[i] = app.PadString(b.Ansi(app.UL, t.GetDate().String()), maxlen, false)
	}

	// Second column -- amount
	amt_cols := make([][]string, len(postingTrans)+1)
	amt_cols[0] = header

	for i := range trans {

		idx := make(map[[3]string]int)
		for tidx, v := range trans[i] {
			idx[[3]string{v.GetAccount(), v.GetAccountTerm(), v.GetCCY()}] = tidx
		}

		for pidx, p := range postingTrans {
			tidx, ok := idx[[3]string{p.GetAccount(), p.GetAccountTerm(), p.GetCCY()}]
			if !ok {
				return fmt.Errorf("account %s (term %s) currency %s not on all transactions; must summarise!",
					p.GetAccount(),
					p.GetAccountTerm(),
					p.GetCCY())
			}
			v := trans[i][tidx]

			if p.GetAccountLevel() != v.GetAccountLevel() || p.GetAccountTerm() != v.GetAccountTerm() {
				return fmt.Errorf("account %s has inconsistent level (%d vs %d) or term (%s vs %s)", p.GetAccount(),
					p.GetAccountLevel(), v.GetAccountLevel(), p.GetAccountTerm(), v.GetAccountTerm())
			}

			if amt_cols[pidx+1] == nil {
				amt_cols[pidx+1] = make([]string, len(trans)+1)
			}
			amt_cols[pidx+1][i] = b.FormatMoney(v.GetCCY(), v.GetAmount(), maxlen)
		}
	}

	lacct := app.ListLength(acct_col, 100)

	// Print final column out
	for i := range acct_col {
		if i > 0 && postingTrans[i-1].GetAccountLevel() == 0 {
			b.Printf("\n")
		}
		b.Printf("%s %s\n",
			app.PadString(acct_col[i], lacct, true),
			strings.Join(amt_cols[i], " "))
	}

	return nil
}
