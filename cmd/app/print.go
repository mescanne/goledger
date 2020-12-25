package app

import (
	"encoding/json"
	"fmt"
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

// Print JSON
func (b *BookPrinter) PrintJSON(v interface{}, pretty bool) error {
	var ob []byte
	var err error
	if pretty {
		ob, err = json.MarshalIndent(v, "", "    ")
	} else {
		ob, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}

	if _, err = b.Write(ob); err != nil {
		return err
	}

	return nil
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
	if amount.Cmp(&zero) >= 0 {
		return b.Ansi(Blue, num)
	} else {
		return b.Ansi(Red, num)
	}
}

func ListLength(strs []string, max int) (l int) {
	if max < 0 {
		max = -1 * max
	}
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

// Combine multiple columns into one.
// Single list length. Right-justified. Space-separated.
func Combine(strs [][]string, max int) ([]string, error) {

	// Calculate width of each column
	maxlen := 0
	for _, str := range strs {
		if len(str) != len(strs[0]) {
			return nil, fmt.Errorf("internal error: inconsistent string lengths")
		}
		l := ListLength(str, max)
		if l > maxlen {
			maxlen = l
		}
	}

	// New column
	ncol := make([]string, len(strs[0]))
	buf := make([]string, len(strs))
	for i := range ncol {
		for j := range strs {
			buf[j] = PadString(strs[j][i], maxlen, false)
		}
		ncol[i] = strings.Join(buf, " ")
	}

	return ncol, nil
}

func Length(s string) int {
	count := 0
	idx := 0
	for idx < len(s) {
		r, width := utf8.DecodeRuneInString(s[idx:])
		idx += width

		// Skip forward if it's ansi
		if r == ansiStart {
			for idx < len(s) && r != ansiEnd {
				r, width = utf8.DecodeRuneInString(s[idx:])
				idx += width
			}

			// Now we have 'm' -- get to the next one
			continue
		}

		// Another character
		count++
	}

	return count
}

const ansiStart = '\033'
const ansiEnd = 'm'

func PadString(s string, max int, justify_left bool) string {

	var sb strings.Builder

	count := 0
	idx := 0
	for idx < len(s) {

		r, width := utf8.DecodeRuneInString(s[idx:])
		idx += width

		// Skip forward if it's ansi
		if r == ansiStart {
			sb.WriteRune(r)

			for idx < len(s) && r != ansiEnd {
				r, width = utf8.DecodeRuneInString(s[idx:])
				idx += width
				sb.WriteRune(r)
			}

			// Now we have 'm' -- get to the next one
			continue
		}

		// Another character
		count++

		if count <= max {
			sb.WriteRune(r)
		}

	}

	if count >= max {
		return sb.String()
	}

	padding := strings.Repeat(" ", max-count)
	if justify_left {
		return sb.String() + padding
	} else {
		return padding + sb.String()
	}
}

func (b *BookPrinter) Ansi(c AnsiColour, i string) string {
	if !b.colour {
		return i
	}
	return string(c) + i + string(Reset)
}

type AnsiColour string

const Blue AnsiColour = "\033[0;34m"
const Red AnsiColour = "\033[0;31m"
const BlueUL AnsiColour = "\033[4;34m"
const BlackUL AnsiColour = "\033[4;30m"
const Black AnsiColour = "\033[0;30m"
const UL AnsiColour = "\033[4m"
const Reset AnsiColour = "\033[0m"
