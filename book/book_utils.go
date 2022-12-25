package book

import (
	"math/big"
	"regexp"
)

// Rename accounts
//
// Replace all accounts matching regular expression 'search' with regular expression (eg using $1) 'replace'
//
// search  search regular expression
// replace replace regular expression string (if non-empty)
// alt     replace string if it doesn't match (if non-empty)
//
func (b *Book) RegexAccounts(search string, replace string, alt string) {
	re := regexp.MustCompile(search)
	b.MapAccount(func(acct string) string {
		if alt != "" && !re.MatchString(acct) {
			return alt
		}
		if replace != "" {
			return re.ReplaceAllString(acct, replace)
		}
		return acct
	})
}

func (b *Book) RegexCCY(search string, replace string) {
	one := big.NewRat(1, 1)
	re := regexp.MustCompile(search)

	// Map all the postings
	b.MapAmount(func(date Date, ccy string) (*big.Rat, string) {
		if re.MatchString(ccy) {
			return one, re.ReplaceAllString(ccy, replace)
		} else {
			return one, ccy
		}
	})

	// Map the prices
	for ccypair, pl := range b.prices.data {
		newpair := ccypair
		if re.MatchString(ccypair.Unit) {
			newpair.Unit = re.ReplaceAllString(ccypair.Unit, replace)
		}
		if re.MatchString(ccypair.CCY) {
			newpair.CCY = re.ReplaceAllString(ccypair.CCY, replace)
		}
		if newpair.Unit == ccypair.Unit && newpair.CCY == ccypair.CCY {
			continue
		}
		b.prices.data[newpair] = pl
		delete(b.prices.data, ccypair)
	}
}

func (b *Book) FilterByDateSince(minDate Date) {
	if minDate == 0 {
		return
	}
	b.FilterTransaction(func(date Date, payee string, posts Transaction) bool {
		return date >= minDate
	})
}

func (b *Book) FilterByDateAsof(maxDate Date) {
	if maxDate == 0 {
		return
	}
	b.FilterTransaction(func(date Date, payee string, posts Transaction) bool {
		return date < maxDate
	})
}

// Filter in all transactions that contain at least one posting matching exactly acct
func (b *Book) FilterByAccount(acct string) {
	b.FilterTransaction(func(date Date, payee string, posts Transaction) bool {
		for _, v := range posts {
			if v.GetAccount() == acct {
				return true
			}
		}
		return false
	})
}
