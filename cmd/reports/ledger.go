package reports

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
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
