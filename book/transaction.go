package book

import (
	"encoding/json"
	"time"
)

type Transaction []Posting

func (t Transaction) MarshalJSON() ([]byte, error) {

	type JsonPosting struct {
		Account string  `json:"account"`
		CCY     string  `json:"ccy"`
		Amount  float64 `json:"amount"`
		Note    string  `json:"note,omitempty"`
	}

	type JsonTransaction struct {
		Date  string        `json:"date"`
		Payee string        `json:"payee,omitempty"`
		Note  string        `json:"note,omitempty"`
		Posts []JsonPosting `json:"posts"`
	}

	var o JsonTransaction
	o.Date = t[0].date.GetTime().Format(time.RFC3339)
	o.Payee = t[0].payee
	o.Note = t[0].tnote
	o.Posts = make([]JsonPosting, len(t), len(t))

	for i, p := range t {
		o.Posts[i].Account = p.acct
		o.Posts[i].CCY = p.ccy
		amt, _ := p.val.Float64()
		o.Posts[i].Amount = amt
		o.Posts[i].Note = p.tnote
	}

	return json.Marshal(&o)
}

func (t Transaction) GetDate() Date {
	return t[0].date
}

func (t Transaction) GetPayee() string {
	return t[0].payee
}

func (t Transaction) GetTransactionNote() string {
	return t[0].tnote
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
