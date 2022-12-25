package export

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/loader"
)

func LedgerFormatCurrency(ccy string) string {
	needsQuotes := false
	for _, c := range ccy {
		if !loader.IsCCYRune(rune(c)) {
			needsQuotes = true
		}
	}
	if needsQuotes {
		ccy = "\"" + ccy + "\""
	}
	return ccy
}

func ShowLedger(b *app.BookPrinter, bk *book.Book) error {
	trans := bk.Transactions()

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

			ccy := LedgerFormatCurrency(p.GetCCY())

			b.Printf("  %s  %s%s%s\n", p.GetAccount(), b.FormatSymbol(ccy), b.FormatNumber(p.GetCCY(), p.GetAmount()), pnote)
		}
		b.Printf("\n")
	}

	b.Printf("\n; Prices\n")
	for _, pair := range bk.GetPricePairs() {
		pl := bk.GetPriceList(pair.Unit, pair.CCY)
		for _, p := range pl {
			b.Printf("P %s 00:00:00 %s %s%s\n", p.GetDate(), pair.Unit, b.FormatSymbol(pair.CCY), b.FormatNumber(pair.CCY, p.GetPrice()))
		}
	}

	return nil
}
