package reports

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"math/big"
	"regexp"
)

func ShowLedger(b *app.BookPrinter, trans []book.Transaction) error {
	for _, posts := range trans {
		b.Printf("%s %s\n", posts.GetDate(), posts.GetPayee())
		for _, p := range posts {
			b.Printf("  %s  %s%s\n", p.GetAccount(), p.GetCCY(), b.FormatNumber(p.GetCCY(), p.GetAmount()))
		}
		b.Printf("\n")
	}
	return nil
}

func ShowTransactions(b *app.BookPrinter, trans []book.Transaction, credit string) error {

	// Compile credit-account matcher
	re, err := regexp.Compile(credit)
	if err != nil {
		return fmt.Errorf("failed compiling credit accounts '%s': %v", credit, err)
	}

	var zero big.Int
	for _, posts := range trans {
		l := posts.MaxAccountTerm(2)
		payee := posts.GetPayee()
		if payee != "" {
			b.Printf("\n%s - %s\n", posts.GetDate(), posts.GetPayee())
		}

		// Find max length for numbers
		maxlen := 0
		for _, v := range posts {
			f, _ := v.GetAmount().Float64()
			nl := len(b.Sprintf("%.0f", f))
			if nl > maxlen {
				maxlen = nl
			}
		}
		maxlen = maxlen + 1

		for _, v := range posts {
			if v.GetAmount().Num().Cmp(&zero) == 0 {
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

			b.Printf("%s\n", b.FormatMoney(v.GetCCY(), &amt, maxlen))
		}
	}

	return nil
}
