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

	if rr.ch == '"' {
		return strings.TrimSpace(rr.parseQuotedString())
	}

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

func (rr *basicReader) parseTransaction() (book.Date, string, string) {
	date := rr.parseDate()
	_ = rr.consumeWS()
	if rr.ch == '*' {
		rr.next()
		_ = rr.consumeWS()
	}
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
	if rr.ch == '!' || rr.ch == '[' {
		rr.next()
	}
	for (rr.ch >= 'a' && rr.ch <= 'z') ||
		(rr.ch >= 'A' && rr.ch <= 'Z') ||
		(rr.ch == '&') ||
		(rr.ch == '\'') ||
		(rr.ch >= '0' && rr.ch <= '9') ||
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

func IsCCYRune(c rune) bool {
	return (c != eof &&
		c != ';' &&
		c != '@' &&
		c != eol &&
		!unicode.IsSpace(c) &&
		c != '-' &&
		c != '.' &&
		(c < '0' || c > '9'))
}

func (rr *basicReader) parseCCYAmt() (string, *big.Rat) {
	if rr.ch == '"' || IsCCYRune(rr.ch) {
		ccy := rr.parseCCY()
		amt := rr.parseAmt()
		return ccy, amt
	} else {
		amt := rr.parseAmt()
		ccy := rr.parseCCY()
		return ccy, amt
	}
}

func (rr *basicReader) parseCCY() string {
	_ = rr.consumeWS()

	if rr.ch == '"' {
		return rr.parseQuotedString()
	}

	var buf bytes.Buffer
	for IsCCYRune(rr.ch) {
		buf.WriteRune(rr.ch)
		rr.next()
	}

	return buf.String()
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
}

func (rr *basicReader) parsePrice(loader TransactionLoader) {
	if rr.ch != 'P' {
		rr.stop("expected 'P', got '%c'", rr.ch)
	}
	rr.next()

	_ = rr.consumeWS()
	date := rr.parseDate()
	_ = rr.consumeWS()
	_ = rr.parseTime()
	unit := rr.parseCCY()
	_ = rr.consumeWS()
	ccy, amt := rr.parseCCYAmt()
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
	return parseFileLocal(loader, filename, nil)
}

func parseFileLocal(loader TransactionLoader, filename string, alias map[string]string) (reterr error) {
	reterr = nil

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	defer func() {
		if msg := recover(); msg != nil {
			reterr = fmt.Errorf("parsing failure: %s: %s", filename, msg)
		}
	}()

	// Create local copy
	nmap := make(map[string]string)
	if alias != nil {
		for k, v := range alias {
			nmap[k] = v
		}
	}

	rr := newRuneReader(bufio.NewReader(file))
	for rr.ch != eof {

		// Move forward to first non-whitespace
		ws := rr.consumeWS()

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
				rr.parsePrice(loader)
				continue
			}

			// Digit -- parse transaction
			if rr.ch >= '0' && rr.ch <= '9' {
				date, payee, note := rr.parseTransaction()
				loader.NewTransaction(book.Date(date), payee, note)
				continue
			}

			command := rr.parseIdentifier()

			// If it's include
			if command == "INCLUDE" {
				ifile := rr.parseToEOL()
				if ifile[0] == '"' && ifile[len(ifile)-1] == '"' {
					ifile = ifile[1 : len(ifile)-1]
				}
				err := parseFileLocal(loader, filepath.Join(filepath.Dir(filename), ifile), nmap)
				if err != nil {
					return err
				}
			} else if command == "ALIAS" {
				shortAcct := rr.parseAccount()
				rr.consumeWS()
				if rr.ch != '=' {
					rr.stop("expected '=' after alias, got '%c'", rr.ch)
				}
				rr.next()
				rr.consumeWS()
				longAcct := rr.parseAccount()
				rr.consumeWS()
				if longAcct == "" {
					rr.stop("expected account for alias, got empty account")
				}
				rr.parseToEOL()
				nmap[shortAcct] = longAcct
			} else {
				rr.stop("expected include or alias")
			}
			continue
		}

		// indented means posting!
		rr.parsePosting(loader, nmap)
	}

	return
}

func (rr *basicReader) parsePosting(loader TransactionLoader, alias map[string]string) {
	_ = rr.consumeWS()
	acct := rr.parseAccount()
	nacct, ok := alias[acct]
	if ok {
		acct = nacct
	}
	_ = rr.consumeWS()
	isNeg := false
	if rr.ch == '-' {
		rr.next()
		isNeg = true
	}
	ccy, dec := rr.parseCCYAmt()
	if isNeg {
		dec.Neg(dec)
	}
	_ = rr.consumeWS()
	if rr.ch == '@' {
		rr.next()
		_ = rr.consumeWS()
		_, _ = rr.parseCCYAmt()
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
