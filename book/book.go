package book

import (
	"fmt"
	"math/big"
	"regexp"
	"sort"
)

type Book struct {
	post   []Posting
	trans  []Transaction
	prices *priceBook
	ccy    map[string]int
}

func (b *Book) Transactions() []Transaction {
	b.compact()
	return b.trans
}

func (b *Book) GetPrice(date Date, unit string, ccy string) *big.Rat {
	return b.prices.getPrice(date, unit, ccy)
}

func (b *Book) Duplicate() *Book {
	newp := make([]Posting, len(b.post), len(b.post))
	copy(newp, b.post)
	newt := make([]Transaction, len(b.trans), len(b.trans))
	copy(newt, b.trans)
	return &Book{newp, newt, b.prices, b.ccy}
}

func (b *Book) GetCCYDecimals() map[string]int {
	return b.ccy
}

func (b *Book) SplitBy(by string) {
	if by == "none" {
		return
	}

	b.MapTransaction(func(date Date, payee string) (Date, string) {
		return date.Floor(by), ""
	})
}

//
// Adjust all amounts by a factor so that they represent one of monthly, daily, quarterly, none, or yearly.
// (None is no adjustment)
//
// This should be done before a CombineAll() to sum all of the amounts.
//
// Example:
//   If the book is for 4 months, adjusting by monthly will divide all amounts by four.
//   After CombineAll() it will give you the monthly average over the four months.
//
func (b *Book) AdjustBy(by string) {
	if by == "none" {
		return
	}

	var minDate, maxDate Date
	isFirst := true
	for _, p := range b.post {
		if isFirst {
			minDate = p.date
			maxDate = p.date
			isFirst = false
			continue
		}
		if p.date < minDate {
			minDate = p.date
		}
		if p.date > maxDate {
			maxDate = p.date
		}
	}

	daysDiff := int64(maxDate.DaysSince(minDate))
	var factor *big.Rat
	if by == "monthly" {
		factor = big.NewRat(365, daysDiff*12)
	} else if by == "daily" {
		factor = big.NewRat(1, daysDiff)
	} else if by == "quarterly" {
		factor = big.NewRat(365, daysDiff*4)
	} else if by == "yearly" {
		factor = big.NewRat(365, daysDiff)
	} else {
		panic(fmt.Sprintf("Invalid time adjustment '%s'", by))
	}

	newposts := make([]Posting, 0, len(b.post))
	for _, p := range b.post {
		newposts = append(newposts, p.byFactor(p.GetAccount(), factor))
	}

	b.post = newposts
}

// Create Adjust Postings
//
// For all postings that match search_acct regular expression, create a pair of postings
// to transfer a portion (by factor) from matching account into replace_acct.
//
// search_acct - posting accounts to match
// replace_acct - account to transfer amount into
// factor - portion of original posting to transform
func (b *Book) AdjustPost(search_acct string, replace_acct string, factor float64) {
	re := regexp.MustCompile(search_acct)
	var r, i big.Rat
	r.SetFloat64(factor)
	i.Neg(&r)
	newposts := b.post
	for _, p := range b.post {
		if re.MatchString(p.GetAccount()) {
			newposts = append(newposts, p.byFactor(p.GetAccount(), &i))
			nacct := re.ReplaceAllString(p.GetAccount(), replace_acct)
			newposts = append(newposts, p.byFactor(nacct, &r))
		}
	}

	b.post = newposts
}

// Find all the accounts matching regular expression reg that have non-zero balances
// TODO: This only applies where balance is already calculated!
func (b *Book) Accounts(reg string) []string {
	re := regexp.MustCompile(reg)
	lbal := make(map[string]*big.Rat)
	accts := make([]string, 0, 10)
	for _, p := range b.post {
		if re.MatchString(p.GetAccount()) {
			lbal[p.GetAccount()] = p.GetBalance()
		}
	}
	var zero big.Int
	for acct, bal := range lbal {
		if bal.Num().Cmp(&zero) != 0 {
			accts = append(accts, acct)
		}
	}
	sort.Strings(accts)
	return accts
}

func (b *Book) FilterTransactionReverse(filter func(date Date, payee string, posts Transaction) bool) {

	// Target
	newp := make(Transaction, 0, len(b.post))
	newt := make([]Transaction, 0, len(b.trans))

	for i := len(b.trans) - 1; i >= 0; i-- {
		p := b.trans[i]
		if filter(p[0].date, p[0].payee, p) {
			idx := len(newp)
			newp = append(newp, p...)
			newt = append(newt, newp[idx:len(newp)])
		}
	}

	b.post = newp
}

func (b *Book) FilterTransaction(filter func(date Date, payee string, posts Transaction) bool) {

	// Target
	newp := make([]Posting, 0, len(b.post))
	newt := make([]Transaction, 0, len(b.trans))

	for _, p := range b.trans {
		if filter(p[0].date, p[0].payee, p) {
			idx := len(newp)
			newp = append(newp, p...)
			newt = append(newt, newp[idx:len(newp)])
		}
	}

	b.post = newp
	b.trans = newt
}

func (b *Book) MapTransaction(mapper func(date Date, payee string) (Date, string)) {
	p := b.post
	for i := range p {
		p[i].date, p[i].payee = mapper(p[i].date, p[i].payee)
	}
}

func (b *Book) MapAmount(mapper func(date Date, ccy string) (*big.Rat, string)) {
	p := b.post
	for i := range p {
		v, ccy := mapper(p[i].date, p[i].ccy)
		p[i].ccy = ccy
		p[i].val.Mul(p[i].val, v)
	}
}

func (b *Book) MapAccount(mapper func(acct string) string) {
	amap := make(map[string]string)
	p := b.post
	for i := range p {
		n, ok := amap[p[i].acct]
		if !ok {
			n = mapper(p[i].acct)
			amap[p[i].acct] = n
		}
		p[i].acct = n
		p[i].acctlevel = 0
		p[i].acctterm = n
	}
}
