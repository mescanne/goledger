package book

import (
	"fmt"
	"math/big"
)

type Posting struct {
	date  Date
	payee string
	tnote string
	acct  string
	ccy   string
	val   *big.Rat
	note  string
	bal   *big.Rat

	// New account levels:
	acctlevel int    // default 0 - no indentation
	acctterm  string // default - same as acct, otherwise term part
	acctsort  string // default "" - sort order for accumulation (intra-transaction)
}

func (p Posting) GetTransactionNote() string { return p.tnote }
func (p Posting) GetAccount() string         { return p.acct }
func (p Posting) GetAccountLevel() int       { return p.acctlevel }
func (p Posting) GetAccountTerm() string     { return p.acctterm }
func (p Posting) GetAmount() *big.Rat        { return p.val }
func (p Posting) GetCCY() string             { return p.ccy }
func (p Posting) GetPostNote() string        { return p.note }
func (p Posting) GetBalance() *big.Rat       { return p.bal }

func (p Posting) byFactor(new_account string, factor *big.Rat) Posting {
	new_post := p
	new_post.acct = new_account
	new_post.acctterm = new_account
	new_post.acctlevel = 0
	new_post.val = big.NewRat(0, 1)
	new_post.val.Mul(p.val, factor)
	new_post.bal = big.NewRat(0, 1)
	return new_post
}

func (p Posting) isLess(r *Posting) bool {

	//
	// Transaction-level sorting
	//
	if p.date < r.date {
		return true
	} else if p.date > r.date {
		return false
	}
	if p.payee < r.payee {
		return true
	} else if p.payee > r.payee {
		return false
	}

	//
	// Posting-level sorting
	//

	if p.acctsort < r.acctsort {
		return true
	} else if p.acctsort > r.acctsort {
		return false
	}

	if p.acct < r.acct {
		return true
	} else if p.acct > r.acct {
		return false
	}

	// Account level
	if p.acctlevel < r.acctlevel {
		return true
	} else if p.acctlevel > r.acctlevel {
		return false
	}

	// Currency
	if p.ccy < r.ccy {
		return true
	}
	return false
}

func (p Posting) String() string {
	pval, _ := p.val.Float64()
	return fmt.Sprintf("Date: %d, Payee: %s, TNote: %s, Acct: %s, CCY: %s, Value: %f, Bal: %v, Note: %s Level: %d Term: %s",
		p.date, p.payee, p.tnote, p.acct, p.ccy, pval, p.bal, p.note, p.acctlevel, p.acctterm)
}
