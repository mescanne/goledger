package loader

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/mescanne/goledger/book"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// TransactLoader - maps a parsed ledger file
// into a series of transactions.
type TransactionLoader interface {
	// Start a new transaction
	NewTransaction(date book.Date, payee string, note string)
	// Add a posting to the current transaction
	AddPosting(acct string, ccy string, amt *big.Rat, note string)
	// Get the current transaction open balances (useful for implicit values)
	GetLastCCYBals() map[string]*big.Rat
	// Add a price for a unit (share price, currency, etc)
	AddPrice(date book.Date, unit string, ccy string, val *big.Rat)
}

type runeReader struct {
	basicReader
}

func (rr *basicReader) parsePayee() string {
	_ = rr.consumeWS()

	var buf bytes.Buffer
	for rr.ch != eof && rr.ch != ';' && rr.ch != eol {
		buf.WriteRune(rr.ch)
		rr.next()
	}
	return strings.TrimSpace(buf.String())
}

func (rr *basicReader) parseNote() string {
	_ = rr.consumeWS()
	if rr.ch != ';' {
		return ""
	} else {
		rr.next()
		_ = rr.consumeWS()
	}

	var buf bytes.Buffer
	for rr.ch != eof && rr.ch != eol {
		buf.WriteRune(rr.ch)
		rr.next()
	}
	return buf.String()
}

//func newRuneReader(b *book.BookBuilder, prices *book.PriceBook, reader *bufio.Reader) *basicReader {
/*
func newRuneReader(reader *bufio.Reader) *runeReader {
  r := &runeReader{
    basicReader: basicReader{
      reader: reader,
    },
    //book: b,
    //prices: prices,
    //trans: nil,
  }
  r.next()
  return r
}
*/

func (rr *basicReader) parseTransaction() (book.Date, string, string) {
	date := rr.parseDate()
	payee := rr.parsePayee()
	note := rr.parseNote()
	if rr.ch == eol {
		rr.next()
	}
	return date, payee, note
}

func (rr *basicReader) parseToEOL() string {
	_ = rr.consumeWS()

	var buf bytes.Buffer
	for rr.ch != eol && rr.ch != eof {
		buf.WriteRune(rr.ch)
		rr.next()
	}

	return buf.String()
}

func (rr *basicReader) parseAccount() string {
	_ = rr.consumeWS()

	var buf bytes.Buffer
	// TODO: Huh? Obscure ledger feature
	if rr.ch == '!' || rr.ch == '[' {
		rr.next()
	}
	for (rr.ch >= 'a' && rr.ch <= 'z') ||
		(rr.ch >= 'A' && rr.ch <= 'Z') ||
		(rr.ch == '&') || // TODO: Remove this
		(rr.ch == '\'') || // TODO: Remove this
		(rr.ch >= '0' && rr.ch <= '9') || // Odd, but valid
		rr.ch == ':' || rr.ch == '_' {
		buf.WriteRune(rr.ch)
		rr.next()
	}
	for (rr.ch >= 'a' && rr.ch <= 'z') ||
		(rr.ch >= 'A' && rr.ch <= 'Z') ||
		rr.ch == ':' || rr.ch == '_' {
		buf.WriteRune(rr.ch)
		rr.next()
	}
	if rr.ch == ']' {
		rr.next()
	}

	return buf.String()
}

func (rr *basicReader) parseCCY() string {
	_ = rr.consumeWS()

	if rr.ch == '"' {
		return rr.parseQuotedString()
	}

	var buf bytes.Buffer
	runes := make([]rune, 0, 10)
	for rr.ch != eof && rr.ch != ';' &&
		rr.ch != eol &&
		!unicode.IsSpace(rr.ch) &&
		rr.ch != '-' && rr.ch != '.' &&
		(rr.ch < '0' || rr.ch > '9') {
		buf.WriteRune(rr.ch)
		runes = append(runes, rr.ch)
		rr.next()
	}

	r := buf.String()
	if r == "£ " {
		fmt.Printf("Got ccy: '%s' from %v\n", r, runes)
	}

	return r
}

func (rr *basicReader) parseInclude() string {
	ident := rr.parseAccount()
	if ident != "include" {
		panic(fmt.Sprintf("Expected 'include', got %s at %c!", ident, rr.ch))
	}

	_ = rr.consumeWS()

	return rr.parseToEOL()
}

func (rr *basicReader) skipLine() {
	var buf bytes.Buffer
	for rr.ch != eof && rr.ch != eol {
		buf.WriteRune(rr.ch)
		rr.next()
	}
	if rr.ch == eol {
		rr.next()
	}
	//log.Printf("Skipped line: %s!\n", buf.String())
}

func (rr *basicReader) parsePrice2(loader TransactionLoader) {
	if rr.ch != 'P' {
		panic(fmt.Sprintf("Expected 'P', got %v", rr.ch))
	}
	rr.next()

	_ = rr.consumeWS()
	date := rr.parseDate()
	_ = rr.consumeWS()
	_ = rr.parseTime()
	unit := rr.parseCCY()
	_ = rr.consumeWS()
	ccy := rr.parseCCY()
	amt := rr.parseAmt()
	loader.AddPrice(book.Date(date), unit, ccy, amt)
	if rr.ch == eol {
		rr.next()
	}
}

// Parse a ledger formatted file (see https://www.ledger-cli.org/) and
// load the transactions and postings into the TransactionLoader
// interface.
//
// This accepts only a subset of the ledger format.
//
// reterr returns the parsing error or nil if everything loaded.
func ParseFile(loader TransactionLoader, filename string) (reterr error) {
	reterr = nil

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lines := 0

	defer func() {
		if msg := recover(); msg != nil {
			reterr = fmt.Errorf("parsing failure: line %d, filename %s, error: %s", (lines - 1), filename, msg)
		}
	}()

	//rr := newRuneReader(prices, bufio.NewReader(file))
	rr := newRuneReader(bufio.NewReader(file))
	for rr.ch != eof {
		lines++

		// Move forward to first non-whitespace
		ws := rr.consumeWS()

		//log.Printf("First character after %v whitespace: %c.. EOL? %v\n", ws, rr.ch, rr.ch == eol)

		// Skip comment, eof
		if rr.ch == ';' || rr.ch == eol || rr.ch == eof {
			rr.skipLine()
			continue
		}

		// No indentation..
		if ws == 0 {

			// D or = -- ignore this for now.
			if rr.ch == '=' || rr.ch == 'D' {
				rr.skipLine()
				continue
			}

			if rr.ch == 'P' {
				rr.parsePrice2(loader)
				continue
			}

			// Digit -- parse transaction
			if rr.ch >= '0' && rr.ch <= '9' {
				date, payee, note := rr.parseTransaction()
				loader.NewTransaction(book.Date(date), payee, note)
				continue
			}

			// Expect "include <filename>"
			ifile := rr.parseInclude()
			// log.Printf("Parsed filename:%v!", ifile)
			ParseFile(loader, filepath.Join(filepath.Dir(filename), ifile))
			continue

			// panic(fmt.Sprintf("Unexpected character: %v", rr.ch))
		}

		// indented means posting!
		rr.parsePosting(loader)
	}

	return
}

func (rr *basicReader) parsePosting(loader TransactionLoader) {
	_ = rr.consumeWS()
	acct := rr.parseAccount()
	_ = rr.consumeWS()
	isNeg := false
	if rr.ch == '-' {
		rr.next()
		isNeg = true
	}
	ccy := rr.parseCCY()
	dec := rr.parseAmt()
	if isNeg {
		dec.Neg(dec)
	}
	_ = rr.consumeWS()
	if rr.ch == '@' {
		rr.next()
		_ = rr.parseCCY()
		_ = rr.parseAmt()
	}
	note := rr.parseNote()
	if rr.ch == eol {
		rr.next()
	}

	if ccy == "" {
		// ASSERT: dec should be 0
		var ZERO big.Int
		for ccy, dec := range loader.GetLastCCYBals() {
			if dec.Num().Cmp(&ZERO) != 0 {
				var x big.Rat
				x.Neg(dec)
				loader.AddPosting(acct, ccy, &x, note)
			}
		}
	} else {
		// Add post
		loader.AddPosting(acct, ccy, dec, note)
	}

	// TODO: Add Pricing!
}
