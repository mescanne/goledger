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
		return b.Blue(num)
	} else {
		return b.Red(num)
	}
}

// Whether Colour is enabled or not
func (b *BookPrinter) Colour() bool {
	return b.colour
}

// Highlight the string as blue (if colour is enabled)
func (b *BookPrinter) Blue(i string) string {
	if !b.colour {
		return i
	}
	return __AnsiBlue + i + __AnsiReset
}

// Highlight the string as red (if colour is enabled)
func (b *BookPrinter) Red(i string) string {
	if !b.colour {
		return i
	}
	return __AnsiRed + i + __AnsiReset
}

// Highlight the string as blue underline (if colour is enabled)
func (b *BookPrinter) BlueUL(i string) string {
	if !b.colour {
		return i
	}
	return __AnsiBlueUL + i + __AnsiReset
}

// Format the number (with colour if enabled) to a maximum length
// (between symbol and number) and return the string
func (b *BookPrinter) FormatMoney(symbol string, amount *big.Rat, maxlen int) string {
	sym := b.FormatSymbol(symbol)
	l := maxlen - utf8.RuneCountInString(sym)
	num := b.pr.Sprintf("%s%*s", sym, l, b.FormatNumber(symbol, amount))
	var zero big.Rat
	if b.colour {
		num = strings.ReplaceAll(num, "  ", " ·")
	}
	if amount.Cmp(&zero) >= 0 {
		return b.Blue(num)
	} else {
		return b.Red(num)
	}
}

const __AnsiReset = "\033[0m"
const __AnsiBlue = "\033[34m"
const __AnsiRed = "\033[31m"
const __AnsiBlueUL = "\033[4;34m"
