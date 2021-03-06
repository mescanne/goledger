package script

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"go.starlark.net/starlark"
)

type starlarkPosting book.Posting

func (p starlarkPosting) Attr(name string) (starlark.Value, error) {
	post := book.Posting(p)
	switch name {
	case "date":
		return starlark.String(post.GetDate().String()), nil
	case "payee":
		return starlark.String(post.GetPayee()), nil
	case "account":
		return starlark.String(post.GetAccount()), nil
	case "transaction_note":
		return starlark.String(post.GetTransactionNote()), nil
	case "amount":
		v, _ := post.GetAmount().Float64()
		return starlark.Float(v), nil
	case "ccy":
		return starlark.String(post.GetCCY()), nil
	case "posting_note":
		return starlark.String(post.GetPostNote()), nil
	case "balance":
		v, _ := post.GetBalance().Float64()
		return starlark.Float(v), nil
	default:
		return nil, nil
	}
}
func (s starlarkPosting) AttrNames() []string {
	return []string{
		"date", "payee", "transaction_note", "account",
		"amount", "ccy", "posting_note", "balance",
	}
}
func (s starlarkPosting) String() string {
	return book.Posting(s).String()
}
func (s starlarkPosting) Type() string {
	return "posting"
}
func (s starlarkPosting) Freeze() {}
func (s starlarkPosting) Truth() starlark.Bool {
	return true
}
func (s starlarkPosting) Hash() (uint32, error) {
	return 0, fmt.Errorf("cannot hash posting")
}

type starlarkTransaction book.Transaction

type starlarkIterator struct {
	starlark.Indexable
	i int
}

func (p *starlarkIterator) Next(v *starlark.Value) bool {
	if p.i == p.Len() {
		return false
	}
	*v = p.Index(p.i)
	p.i += 1
	return true
}
func (p *starlarkIterator) Done() {
}
func (s starlarkTransaction) Iterate() starlark.Iterator {
	return &starlarkIterator{s, 0}
}
func (s starlarkTransaction) Index(i int) starlark.Value {
	return starlarkPosting(s[i])
}
func (s starlarkTransaction) Len() int {
	return len(s)
}
func (s starlarkTransaction) Attr(name string) (starlark.Value, error) {
	trans := book.Transaction(s)
	switch name {
	case "date":
		return starlark.String(trans.GetDate().String()), nil
	case "payee":
		return starlark.String(trans.GetPayee()), nil
	case "transaction_note":
		return starlark.String(trans.GetTransactionNote()), nil
	default:
		return nil, nil
	}
}

func (s starlarkTransaction) AttrNames() []string {
	return []string{
		"date", "payee", "transaction_note",
	}
}

func (s starlarkTransaction) String() string {
	return fmt.Sprintf("transaction[%d postings]", len(s))
}
func (s starlarkTransaction) Type() string {
	return "transaction"
}
func (s starlarkTransaction) Freeze() {}
func (s starlarkTransaction) Truth() starlark.Bool {
	return true
}
func (s starlarkTransaction) Hash() (uint32, error) {
	return 0, fmt.Errorf("cannot hash transaction")
}

type starlarkTransactions []book.Transaction

func (s starlarkTransactions) Iterate() starlark.Iterator {
	return &starlarkIterator{s, 0}
}
func (s starlarkTransactions) Index(i int) starlark.Value {
	return starlarkTransaction(s[i])
}
func (s starlarkTransactions) Len() int {
	return len(s)
}

func (s starlarkTransactions) String() string {
	return fmt.Sprintf("book[%d transactions]", len(s))
}
func (s starlarkTransactions) Type() string {
	return "transactions"
}
func (s starlarkTransactions) Freeze() {}
func (s starlarkTransactions) Truth() starlark.Bool {
	return true
}
func (s starlarkTransactions) Hash() (uint32, error) {
	return 0, fmt.Errorf("cannot hash transactions")
}
