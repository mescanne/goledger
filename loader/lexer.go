package loader

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/mescanne/goledger/book"
	"io"
	"log"
	"math/big"
	"unicode"
)

type basicReader struct {
	reader *bufio.Reader
	ch     rune
	row    int
	col    int
}

const eof = rune(0)
const eol = rune('\n')

func newRuneReader(r *bufio.Reader) *basicReader {
	rr := &basicReader{
		reader: r,
		row:    1,
		col:    0,
	}
	rr.next()
	return rr
}

func (rr *basicReader) stop(msgf string, args ...interface{}) {
	pos := fmt.Sprintf("%d:%d: ", rr.row, rr.col)
	msg := fmt.Sprintf(msgf, args...)
	panic(pos + msg)
}

func (rr *basicReader) parseIdentifier() string {
	_ = rr.consumeWS()

	var buf bytes.Buffer
	for unicode.IsDigit(rr.ch) ||
		unicode.IsLetter(rr.ch) ||
		rr.ch == '_' {
		buf.WriteRune(unicode.ToUpper(rr.ch))
		rr.next()
	}

	rr.consumeWS()

	return buf.String()
}

func (rr *basicReader) parseNumber(v *int, min int, max int) int {
	chars := 0
	var retval int = 0
	for rr.ch >= '0' && rr.ch <= '9' {
		retval *= 10
		retval += int(rr.ch - rune('0'))
		chars++
		rr.next()
	}
	if chars < min || chars > max {
		rr.stop("expected number digit length of min:max of %d:%d, found %d digits", min, max, chars)
	}
	return retval
}

func (rr *basicReader) parseDate() book.Date {
	var date int
	date = rr.parseNumber(&date, 1, 4)
	rr.consume('/')
	date = (date * 100) + rr.parseNumber(&date, 1, 2)
	rr.consume('/')
	date = (date * 100) + rr.parseNumber(&date, 1, 2)
	return book.Date(date)
}

func (rr *basicReader) parseTime() int {
	var time int
	time = rr.parseNumber(&time, 1, 2)
	rr.consume(':')
	time = (time * 100) + rr.parseNumber(&time, 1, 2)
	rr.consume(':')
	time = (time * 100) + rr.parseNumber(&time, 1, 2)
	return time
}

func (rr *basicReader) consumeWS() int {
	var consumed int = 0
	//for rr.ch == ' ' || rr.ch == '\t' {
	for unicode.IsSpace(rr.ch) && rr.ch != '\n' {
		consumed++
		rr.next()
	}
	return consumed
}

func (rr *basicReader) consume(e rune) {
	if rr.ch != e {
		rr.stop("expected '%c', got '%c'", e, rr.ch)
	}
	rr.next()
}

func (rr *basicReader) next() rune {
	r, _, err := rr.reader.ReadRune()
	if err == io.EOF {
		r = eof
	} else if err != nil {
		log.Fatal(err)
	}
	rr.ch = r
	if rr.ch == eol {
		rr.row++
		rr.col = 0
	} else {
		rr.col++
	}
	return r
}

func (rr *basicReader) parseQuotedString() string {
	rr.consume('"')

	var buf bytes.Buffer
	for rr.ch != eof && rr.ch != eol && rr.ch != '"' {
		buf.WriteRune(rr.ch)
		rr.next()
	}

	rr.consume('"')

	return buf.String()
}

func (rr *basicReader) parseAmt() *big.Rat {
	_ = rr.consumeWS()

	isNeg := false
	if rr.ch == '-' {
		isNeg = true
		rr.next()
	}

	isDec := false
	var num int64 = 0
	var mag int64 = 1
	for {
		if rr.ch >= '0' && rr.ch <= '9' {
			num *= 10
			num += int64(rr.ch - '0')
			if isDec {
				mag *= 10
			}
			rr.next()
			continue
		}
		if !isDec && rr.ch == '.' {
			isDec = true
			rr.next()
			continue
		}
		if rr.ch == ',' {
			rr.next()
			continue
		}

		break
	}

	if isNeg {
		num *= -1
	}

	return big.NewRat(num, mag)
}
