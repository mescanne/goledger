package book

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
