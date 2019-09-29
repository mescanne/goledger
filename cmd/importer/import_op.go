package importer

import (
	"encoding/csv"
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/utils"
	"io"
	"math/big"
)

type Import struct {
	Payee  string
	Date   book.Date
	Amount *big.Rat
}

type BookImporter func(r io.Reader) ([]Import, error)

var ImportFormatUsage = `Import Formats

The format for the import configuration:

  type:key=value[,key=value,...]

There is one import format type 'csv', but more can be added in the future.

Import type 'csv' Parameters

  payee, date, amount - 0-based column index for the payee, transaction date,
                        and transaction amount
  skip                - number of header lines to skip (default is 0)
  delim               - delimiter for CSV file (default is ,)

`

func NewBookImporterByConfig(cfg *utils.CLIConfig) (BookImporter, error) {
	if cfg.ConfigType == "" {
		return nil, fmt.Errorf("missing import type")
	} else if cfg.ConfigType == "csv" {
		return NewCSVBookImporter(cfg)
	} else {
		return nil, fmt.Errorf("invalid import type: %s", cfg.ConfigType)
	}
}

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

	skip := cfg.GetIntDefault("skip", 0)

	return func(r io.Reader) ([]Import, error) {
		csvr := csv.NewReader(r)
		csvr.Comma = []rune(delim)[0]
		csvr.TrimLeadingSpace = true
		csvr.ReuseRecord = true
		output := make([]Import, 0, 100)
		skiprec := skip
		for {
			recs, err := csvr.Read()
			if err == io.EOF {
				break
			}
			if skiprec > 0 {
				skiprec -= 1
				continue
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
