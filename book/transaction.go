package book

import (
	"encoding/json"
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
