package app

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"io"
	"math/big"
	"strings"
	"unicode"
	"unicode/utf8"
)

// BookPrinter provides formatting for console-based
// accounting reports
type BookPrinter struct {
	w      io.Writer
	pr     *message.Printer
	decs   map[string]int
	colour bool
}

// Create a new BookPrinter for a specified io.Writer and
// decimal CCY (symbol) formatting.
//
// Normally decs comes from GetCCYDecimal() from a book.
func (app *App) NewBookPrinter(w io.Writer, decs map[string]int) *BookPrinter {

	// Create a printer for a number of languages
	c := catalog.NewBuilder(catalog.Fallback(language.English))
	c.SetString(language.Dutch, "", "")
	c.SetString(language.English, "", "")
	c.SetString(language.German, "", "")
	c.SetString(language.French, "", "")
	message.DefaultCatalog = c

	// New printer for those languages
	pr := message.NewPrinter(message.MatchLanguage(app.Lang))

	return &BookPrinter{
		w:      w,
		pr:     pr,
		decs:   decs,
		colour: app.Colour,
	}
}

func (b *BookPrinter) Write(p []byte) (int, error) {
	return b.w.Write(p)
}

// Normal Printf() but to specified io.Writier and formatting
// numbers in a locale-specific way
func (b *BookPrinter) Printf(format string, a ...interface{}) (n int, err error) {
	return b.pr.Fprintf(b.w, format, a...)
}

// Normal Sprintf() but formatting numbers in a locale-specific way
func (b *BookPrinter) Sprintf(format string, a ...interface{}) string {
	return b.pr.Sprintf(format, a...)
}

// Format the number correctly based on the symbol in a locale-specific
// way
func (b *BookPrinter) FormatNumber(symbol string, amount *big.Rat) string {
	f, _ := amount.Float64()
	dec, ok := b.decs[symbol]
	if !ok {
		dec = 2
	}
	return b.pr.Sprintf("%.*f", dec, f)
}

// Formal the symbol (CCY) for printing
func (b *BookPrinter) FormatSymbol(symbol string) string {
	if sym := []rune(symbol); unicode.IsLetter(sym[len(sym)-1]) {
		return symbol + " "
	}
	return symbol
}

// Format money in a locale-specific way with the symbol
func (b *BookPrinter) FormatSimpleMoney(symbol string, amount *big.Rat) string {
	num := b.FormatSymbol(symbol) + b.FormatNumber(symbol, amount)
	var zero big.Rat
	if amount.Cmp(&zero) >= 0 {
		return b.Ansi(Blue, num)
	} else {
		return b.Ansi(Red, num)
	}
}

// Format the number (with colour if enabled) to a maximum length
// (between symbol and number) and return the string
func (b *BookPrinter) FormatMoney(symbol string, amount *big.Rat, maxlen int) string {
	sym := b.FormatSymbol(symbol)
	l := maxlen - utf8.RuneCountInString(sym)
	num := b.pr.Sprintf("%s%*s", sym, l, b.FormatNumber(symbol, amount))
	var zero big.Rat
	if b.colour {
		num = strings.ReplaceAll(num, "  ", " \u00B7")
	}
	if amount.Cmp(&zero) >= 0 {
		return b.Ansi(Blue, num)
	} else {
		return b.Ansi(Red, num)
	}
}

func Length(s string) int {
	return utf8.RuneCountInString(s)
}

func ListLength(strs []string, max int) (l int) {
	for _, s := range strs {
		ls := Length(s)
		if ls >= max {
			return max
		}
		if ls > l {
			l = ls
		}
	}
	return
}

func (b *BookPrinter) Ansi(c AnsiColour, i string) string {
	if !b.colour {
		return i
	}
	return string(c) + i + "\033[0m"
}

type AnsiColour string

const Blue AnsiColour = "\033[0;34m"
const Red AnsiColour = "\033[0;31m"
const BlueUL AnsiColour = "\033[4;34m"
const BlackUL AnsiColour = "\033[4;30m"
const Black AnsiColour = "\033[0;30m"
