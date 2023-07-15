package export

import (
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	"encoding/csv"
	"fmt"
)

func ShowCSV(b *app.BookPrinter, bk *book.Book) error {
	csvwrite := csv.NewWriter(b)

	// Header
	hdrs := []string{
		"date",
		"payee",
		"transaction_note",
		"account",
		"currency",
		"amount",
		"post_note",
	}
	csvwrite.Write(hdrs)

	// Postings
	for _, posts := range bk.Transactions() {
		for _, p := range posts {
        		f, _ := p.GetAmount().Float64()
			csvwrite.Write([]string{
				posts.GetDate().String(),
				posts.GetPayee(),
				posts.GetTransactionNote(),
				p.GetAccount(),
				p.GetCCY(),
				fmt.Sprintf("%f", f),
				p.GetPostNote(),
			})
		}
	}

	csvwrite.Flush()

	return nil
}
