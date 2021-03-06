package book

import (
	"fmt"
	"math/big"
)

type Builder struct {
	post      []Posting
	prevTrans map[string]int
	currAmts  map[string]*big.Rat
	currDate  Date
	currPayee string
	currNote  string
	prices    *priceBookBuilder
}

func (b *Builder) Build() *Book {
	b.checkAndClearTransaction()

	// Get the largest denominators by currency
	mmap := make(map[string]*big.Int)
	for _, p := range b.post {
		eval, ok := mmap[p.GetCCY()]
		if !ok || eval.Cmp(p.GetAmount().Denom()) < 0 {
			mmap[p.GetCCY()] = p.GetAmount().Denom()
		}
	}

	// Convert to number of decimals
	rmap := make(map[string]int)
	ten := big.NewInt(10)
	for k, v := range mmap {
		p := big.NewInt(1)
		c := 0
		// for as long as Denom is bigger keep doing p = p * 10
		for p.Cmp(v) < 0 {
			p.Mul(p, ten)
			c += 1
		}
		rmap[k] = c
	}

	nbook := &Book{
		post:   b.post,
		trans:  make([]Transaction, len(b.post), len(b.post)),
		prices: b.prices.build(),
		ccy:    rmap,
	}

	// Compact the book
	nbook.compact()

	return nbook
}

func NewBookBuilder() *Builder {
	return &Builder{
		post:      make([]Posting, 0, 200),
		prevTrans: make(map[string]int),
		currAmts:  make(map[string]*big.Rat),
		currDate:  Date(-1),
		currPayee: "",
		currNote:  "",
		prices:    newPriceBookBuilder(),
	}
}

func (b *Builder) checkAndClearTransaction() {
	var Zero big.Int
	for ccy, amt := range b.currAmts {
		if amt.Num().Cmp(&Zero) != 0 {
			panic(fmt.Sprintf("In transaction (%d %s) ccy %s has balance of %s!", b.currDate, b.currPayee, ccy, amt))
		} else {
			b.currAmts[ccy].SetInt64(0)
		}
	}
}

func (b *Builder) NewTransaction(date Date, payee string, note string) {

	// Check and clear current transaction
	b.checkAndClearTransaction()

	// Validate uniqueness of transaction
	key := fmt.Sprintf("%d %s", date, payee)
	idx, ok := b.prevTrans[key]
	if !ok {
		b.prevTrans[key] = 1
	} else {
		idx++
		b.prevTrans[key] = idx
		payee = fmt.Sprintf("%s (%d)", payee, idx)
	}

	// Initialize new values
	b.currDate = date
	b.currPayee = payee
	b.currNote = note
}

func (b *Builder) AddPosting(acct string, ccy string, amt *big.Rat, note string) {
	b.post = append(b.post, Posting{
		date:  b.currDate,
		payee: b.currPayee,
		tnote: b.currNote,
		acct:  acct,
		ccy:   ccy,
		val:   amt,
		note:  note,
		bal:   big.NewRat(0, 1),
	})

	v, ok := b.currAmts[ccy]
	if ok {
		v.Add(v, amt)
	} else {
		v = big.NewRat(0, 1)
		v.Set(amt)
		b.currAmts[ccy] = v
	}
}

func (b *Builder) GetLastCCYBals() map[string]*big.Rat {
	return b.currAmts
}

func (b *Builder) AddPrice(date Date, unit string, ccy string, val *big.Rat) {
	b.prices.addPrice(date, unit, ccy, val)
}
