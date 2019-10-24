package reports

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"regexp"
	"unicode/utf8"
)

func ShowLedger(b *app.BookPrinter, trans []book.Transaction) error {
	for _, posts := range trans {
		tnote := posts.GetTransactionNote()
		if tnote != "" {
			tnote = "  ; " + tnote
		}
		b.Printf("%s %s%s\n", posts.GetDate(), posts.GetPayee(), tnote)
		for _, p := range posts {
			pnote := p.GetPostNote()
			if pnote != "" {
				pnote = "  ; " + pnote
			}
			b.Printf("  %s  %s%s%s\n", p.GetAccount(), p.GetCCY(), b.FormatNumber(p.GetCCY(), p.GetAmount()), pnote)
		}
		b.Printf("\n")
	}
	return nil
}

func ShowTransactions(b *app.BookPrinter, trans []book.Transaction, credit string, hidden string) error {

	// Compile credit-account matcher
	re, err := regexp.Compile(credit)
	if err != nil {
		return fmt.Errorf("failed compiling credit accounts '%s': %v", credit, err)
	}

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

			// Skip if hidden
			if v.GetAccount() == hidden {
				continue
			}

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

			// Reverse if matching credit
			amt := *v.GetAmount()
			if re.MatchString(v.GetAccount()) {
				amt.Neg(&amt)
			}

			// Print out money
			b.Printf("%s\n", b.FormatMoney(v.GetCCY(), &amt, maxlen))
		}
	}

	return nil
}
