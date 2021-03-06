package book

import (
	"encoding/json"
	"math/big"
	"time"
)

type Transaction []Posting

func (t Transaction) GetDate() Date {
	return t[0].date
}

func (t Transaction) GetPayee() string {
	return t[0].payee
}

func (t Transaction) GetTransactionNote() string {
	return t[0].tnote
}

func (t Transaction) InferRates(base string) map[string]*big.Rat {
	rates := make(map[string]*big.Rat)
	lastAccount := 0
	i := 1
	for {
		// Keep searching for the end of the account set
		if i < len(t) && t[lastAccount].acct == t[i].acct {
			i++
			continue
		}

		// Check if it's precisely 2
		if (i - lastAccount) == 2 {
			// Infer a rate
			rate := big.NewRat(0, 1)
			if t[i-1].ccy == base {
				rate.Quo(t[i-1].val, t[i-2].val)
				rate.Neg(rate)
				rates[t[i-2].ccy] = rate
			} else if t[i-2].ccy == base {
				rate.Quo(t[i-2].val, t[i-1].val)
				rate.Neg(rate)
				rates[t[i-1].ccy] = rate
			}
		}

		lastAccount = i
		i++
	}

	return rates
}

func (t Transaction) InferRate(base string, unit string) *big.Rat {
	lastAccount := -1
	i := 0
	for {
		// Check if two consecutive match accounts
		// and are in opposite directions
		if (i-lastAccount) == 2 && (t[i-1].val.Sign()+t[i-2].val.Sign()) == 0 {

			// Extract rate if they match base /unit
			rate := big.NewRat(0, 1)
			if t[i-1].ccy == base && t[i-2].ccy == unit {
				rate.Quo(t[i-1].val, t[i-2].val)
				rate.Neg(rate)
				return rate
			} else if t[i-1].ccy == unit && t[i-2].ccy == base {
				rate.Quo(t[i-2].val, t[i-1].val)
				rate.Neg(rate)
				return rate
			}
		}

		// If at end, break
		if i == len(t) {
			break
		}

		// Reset last account
		if lastAccount == -1 || t[lastAccount].acct != t[i].acct {
			lastAccount = i
		}

		i++
	}

	return nil
}

func (t Transaction) MaxAccountTerm(paddingByLevel int) int {
	m := 0
	for _, p := range t {
		tlen := len(p.acctterm) + p.acctlevel*paddingByLevel
		if tlen > m {
			m = tlen
		}
	}
	return m
}

func (t Transaction) MarshalJSON() ([]byte, error) {

	type JsonTransaction struct {
		Date  string    `json:"date"`
		Payee string    `json:"payee,omitempty"`
		Note  string    `json:"note,omitempty"`
		Posts []Posting `json:"posts"`
	}

	return json.Marshal(&JsonTransaction{
		Date:  t[0].date.GetTime().Format(time.RFC3339),
		Payee: t[0].payee,
		Note:  t[0].tnote,
		Posts: t,
	})
}
