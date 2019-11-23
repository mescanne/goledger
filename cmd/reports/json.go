package reports

import (
	"encoding/json"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
)

func ShowJsonLedger(b *app.BookPrinter, trans []book.Transaction, pretty bool) error {
	var ob []byte
	var err error
	if pretty {
		ob, err = json.MarshalIndent(&trans, "", "    ")
	} else {
		ob, err = json.Marshal(&trans)
	}
	if err != nil {
		return err
	}

	if _, err = b.Write(ob); err != nil {
		return err
	}

	return nil
}
