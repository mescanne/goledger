package export

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

func ShowBeancount(b *app.BookPrinter, bk *book.Book, baseCCY string) error {

	trans := bk.Transactions()

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

	// Operating currency
	if baseCCY != "" {
		b.Printf("option \"operating_currency\" \"%s\"\n\n", baseCCY)
	}

	// Account open close
	for acct, stats := range lstats {
		b.Printf("%s open %s\n", stats.first, acct)
		if stats.isZero {
			b.Printf("%s close %s\n", stats.last.AddDays(1), acct)
		}
	}

	// Transactions
	b.Printf("\n; Transactions\n")
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

	// Prices
	b.Printf("\n; Prices\n")
	for _, pair := range bk.GetPricePairs() {
		pl := bk.GetPriceList(pair.Unit, pair.CCY)
		unit := FormatCurrency(pair.Unit)
		ccy := FormatCurrency(pair.CCY)
		for _, p := range pl {
			b.Printf("P %s 00:00:00 %s %s %s\n", p.GetDate(), unit, b.FormatNumber(pair.CCY, p.GetPrice()), ccy)
		}
	}

	return nil
}
