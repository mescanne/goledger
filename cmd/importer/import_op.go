package importer

import (
	"encoding/csv"
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/utils"
	"io"
	"math/big"
)

var ImportFormatUsage = `Import Formats

Description here.

`

// TODO: Update description per section
// TODO: BookImporter becomes a reader -> array of (payee, date, amount) tuple

type Import struct {
	Payee  string
	Date   book.Date
	Amount *big.Rat
}

type BookImporter func(r io.Reader) ([]Import, error)

//type BookImporter func(r io.Reader) (*book.Book, error)

func NewCSVBookImporter(cfg *utils.CLIConfig) (BookImporter, error) {

	payee_col, err := cfg.GetInt("payee")
	if err != nil {
		return nil, err
	}
	date_col, err := cfg.GetInt("date")
	if err != nil {
		return nil, err
	}
	amount_col, err := cfg.GetInt("amount")
	if err != nil {
		return nil, err
	}
	maxcol := payee_col
	if date_col > maxcol {
		maxcol = date_col
	}
	if amount_col > maxcol {
		maxcol = amount_col
	}

	delim := cfg.GetStringDefault("delim", ",")
	if len([]rune(delim)) != 1 {
		return nil, fmt.Errorf("invalid delimiter '%s': length not one character", delim)
	}

	return func(r io.Reader) ([]Import, error) {
		csvr := csv.NewReader(r)
		csvr.Comma = []rune(delim)[0]
		csvr.TrimLeadingSpace = true
		csvr.ReuseRecord = true
		output := make([]Import, 0, 100)
		for {
			recs, err := csvr.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("error importing csv: %v", err)
			}
			if maxcol >= len(recs) {
				return nil, fmt.Errorf("error importing csv: need at least %d columns, and found only %d: %v", maxcol+1, len(recs)+1, recs)
			}
			date := book.DateFromString(recs[date_col])
			if date == book.Date(0) {
				return nil, fmt.Errorf("invalid date %s", recs[date_col])
			}
			amount, ok := big.NewRat(0, 1).SetString(recs[amount_col])
			if !ok {
				return nil, fmt.Errorf("invalid amount %s", recs[amount_col])
			}
			payee := recs[payee_col]
			output = append(output, Import{
				Payee:  payee,
				Date:   date,
				Amount: amount,
			})
		}
		return output, nil
	}, nil
}

//type BookImporter func(r io.Reader) (*book.Book, error)

func NewBookImporterByConfig(cfg *utils.CLIConfig) (BookImporter, error) {
	if cfg.ConfigType == "" {
		return nil, fmt.Errorf("missing import type")
	} else if cfg.ConfigType == "csv" {
		return NewCSVBookImporter(cfg)
	} else {
		return nil, fmt.Errorf("invalid import type: %s", cfg.ConfigType)
	}
}
