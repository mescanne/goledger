package reports

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"unicode/utf8"
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

		// Find max length for account names
		l := posts.MaxAccountTerm(2)

		// Find max length for numbers
		maxlen := 0
		for _, v := range posts {
			rv := b.FormatNumber(v.GetCCY(), v.GetAmount())
			nl := utf8.RuneCountInString(v.GetCCY()) + len(rv) + 1
			if nl > maxlen {
				maxlen = nl
			}
		}

		for _, v := range posts {

			// Print out account levels
			parts := v.GetAccountLevel()
			term := v.GetAccountTerm()
			if parts == 0 {
				if b.Colour() {
					b.Printf("\n")
				}
				b.Printf("%s", b.BlueUL(b.Sprintf("%-*.*s    ", l, l, term)))
			} else {
				for i := 0; i < parts; i++ {
					b.Printf("  ")
				}
				b.Printf("%-*.*s    ", l, l, term)
			}

			// Print out money
			b.Printf("%s\n", b.FormatMoney(v.GetCCY(), v.GetAmount(), maxlen))
		}
	}

	return nil
}
