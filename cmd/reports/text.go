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

	fmts := make([]bool, len(trans)+1)
	fmts[0] = false
	for i := range trans {
		fmts[i+1] = false
	}

	// Header
	header := make([]app.ColumnValue, len(trans)+1)
	header[0] = app.ColumnString(b.Ansi(app.UL, "Account"))
	for i, t := range trans {
		header[i+1] = app.ColumnRightString(b.Ansi(app.UL, t.GetDate().String()))
	}

	// Content
	rows := make([][]app.ColumnValue, 0, len(postingTrans)+1+10)
	rows = append(rows, header)

	// First column and rows initialize
	idx := make(map[[3]string]int)
	for i, v := range postingTrans {

		// Get account name
		lvl := v.GetAccountLevel()
		var t string
		if lvl == 0 {
			t = b.Ansi(app.BlueUL, v.GetAccountTerm())
		} else {
			acctterm := v.GetAccountTerm()
			t = strings.Repeat("  ", lvl) + acctterm
		}

		// Insert newline if needed
		if i > 0 && v.GetAccountLevel() == 0 {
			rows = append(rows, nil)
		}

		pidx := len(rows)
		rows = append(rows, make([]app.ColumnValue, len(trans)+1))
		rows[pidx][0] = app.ColumnString(t)

		// Record the index
		idx[[3]string{v.GetAccount(), v.GetAccountTerm(), v.GetCCY()}] = pidx
	}

	// Now fill in the data
	for i := range trans {

		for _, v := range trans[i] {

			// Find keys for posting
			pidx, ok := idx[[3]string{v.GetAccount(), v.GetAccountTerm(), v.GetCCY()}]
			if !ok {
				return fmt.Errorf("account %s (term %s) currency %s not on all transactions; must summarise!",
					v.GetAccount(),
					v.GetAccountTerm(),
					v.GetCCY())
			}

			// Record amount
			rows[pidx][i+1] = b.GetColumnMoney(v.GetCCY(), v.GetAmount())
		}
	}

	b.PrintColumns(rows, fmts)

	return nil
}
