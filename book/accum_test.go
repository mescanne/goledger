package book

import (
	"fmt"
	"math/big"
	"testing"
)

//
// Input into making a real book
//
type QuickPrice struct {
	unit string
	ccy  string
	rate int64
}

type QuickPosting struct {
	acct string
	ccy  string
	amt  int64
}

type QuickBook struct {
	date  string
	payee string
	posts []QuickPosting
}

func GetBook(qb []QuickBook, qp []QuickPrice) *Book {
	b := NewBookBuilder()
	for _, t := range qb {
		b.NewTransaction(DateFromString(t.date), t.payee, "")
		for _, p := range t.posts {
			b.AddPosting(p.acct, p.ccy, big.NewRat(p.amt, 1), "")
		}
	}

	bdate := DateFromString(qb[0].date)
	for _, p := range qp {
		b.AddPrice(bdate, p.unit, p.ccy, big.NewRat(p.rate, 1))
	}

	return b.Build()
}

func DumpQuickBook(t *testing.T, qb []QuickBook) {
	cbook := 0
	cpost := 0
	for {
		t.Logf(" Trans %d, %v post: %v\n", cbook, qb[cbook].date, qb[cbook].posts[cpost])
		cpost++
		if cpost == len(qb[cbook].posts) {
			cpost = 0
			cbook++
		}
		if cbook == len(qb) {
			break
		}
	}
}

//
// Expected output
//
type QuickAccumPosting struct {
	acct string
	ccy  string
	amt  int64
	lvls []string
}

type QuickAccumBook struct {
	date  string
	payee string
	posts []QuickAccumPosting
}

func DumpQuickAccumBook(t *testing.T, qb []QuickAccumBook) {
	cbook := 0
	cpost := 0
	for {
		t.Logf(" Trans %d, %v post: %v\n", cbook, qb[cbook].date, qb[cbook].posts[cpost])
		cpost++
		if cpost == len(qb[cbook].posts) {
			cpost = 0
			cbook++
		}
		if cbook == len(qb) {
			break
		}
	}
}

func ValidBook(t *testing.T, in []QuickBook, p []QuickPrice, ccy string, qb []QuickAccumBook) {

	nb := GetBook(in, p).Accumulate(ccy, ":")

	Errorf := func(err string, args ...interface{}) {
		t.Logf("Error to ccy %s: %s\n", ccy, fmt.Sprintf(err, args...))
		t.Logf("Incoming postings =\n")
		DumpQuickBook(t, in)
		t.Logf("Accumulated =\n")
		for i, posts := range nb {
			for _, post := range posts {
				t.Logf(" Trans %d Post: %v\n", i, post)
			}
		}
		t.Logf("Accumulated expected =\n")
		DumpQuickAccumBook(t, qb)
		t.Fatalf("TEST FAILED!")
	}

	if len(qb) != len(nb) {
		Errorf("Different number of accumulated transactions: expected %d, got %d\n", len(qb), len(nb))
	}

	for cbook, posts := range nb {
		edate := DateFromString(qb[cbook].date)
		if posts.GetDate() != edate {
			Errorf("Expected date %s, got %s\n", edate, posts.GetDate())
		}
		if posts.GetPayee() != qb[cbook].payee {
			Errorf("Expected payee %s, got %s\n", qb[cbook].payee, posts.GetPayee())
		}
		if len(posts) != len(qb[cbook].posts) {
			Errorf("Different number of postings: expected %d, got %d\n", len(qb[cbook].posts), len(posts))
		}
		for cpost, p := range posts {
			qpost := qb[cbook].posts[cpost]
			if p.GetAccount() != qpost.acct {
				Errorf("Expected account %s, got %s in %v\n", qpost.acct, p.GetAccount(), p)
			}
			if p.GetCCY() != qpost.ccy {
				Errorf("Expected account %s, got %s in %v\n", qpost.ccy, p.GetCCY(), p)
			}
			if p.GetAccountLevel() != len(qpost.lvls)-1 {
				Errorf("Expected levels %d, got %d in %v\n", len(qpost.lvls), p.GetAccountLevel(), p)
			}
			if p.GetAccountTerm() != qpost.lvls[len(qpost.lvls)-1] {
				Errorf("Expected account term %s, got %s in %v\n", qpost.lvls[len(qpost.lvls)-1], p.GetAccountTerm(), p)
			}

			eamt := big.NewRat(qpost.amt, 1)
			if p.GetAmount().Cmp(eamt) != 0 {
				Errorf("Expected amount %v, got %v in %v\n", eamt, p.GetAmount(), p)
			}
		}
	}
}

func TestAccum(t *testing.T) {

	// Test accumulation
	ValidBook(t,
		[]QuickBook{
			QuickBook{"2010-10-01", "rand", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy", 100},
				QuickPosting{"Income:T1:B", "ccy", 100},
				QuickPosting{"Income:C", "ccy", 100},
				QuickPosting{"Expense:C", "ccy", -300},
			}},
		},
		[]QuickPrice{},
		"ccy",
		[]QuickAccumBook{
			QuickAccumBook{"2010-10-01", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy", -300, []string{"Expense:C"}},
				QuickAccumPosting{"Income", "ccy", 300, []string{"Income"}},
				QuickAccumPosting{"Income:C", "ccy", 100, []string{"Income", "C"}},
				QuickAccumPosting{"Income:T1", "ccy", 200, []string{"Income", "T1"}},
				QuickAccumPosting{"Income:T1:A", "ccy", 100, []string{"Income", "T1", "A"}},
				QuickAccumPosting{"Income:T1:B", "ccy", 100, []string{"Income", "T1", "B"}},
			}},
		},
	)

	// Same thing a few times over
	ValidBook(t,
		[]QuickBook{
			QuickBook{"2010-10-01", "rand", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy", 100},
				QuickPosting{"Income:T1:B", "ccy", 100},
				QuickPosting{"Income:C", "ccy", 100},
				QuickPosting{"Expense:C", "ccy", -300},
			}},
			QuickBook{"2010-10-02", "rand", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy", 100},
				QuickPosting{"Income:T1:B", "ccy", 100},
				QuickPosting{"Income:C", "ccy", 100},
				QuickPosting{"Expense:C", "ccy", -300},
			}},
			QuickBook{"2010-10-02", "rand2", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy", 100},
				QuickPosting{"Income:T1:B", "ccy", 100},
				QuickPosting{"Income:C", "ccy", 100},
				QuickPosting{"Expense:C", "ccy", -300},
			}},
		},
		[]QuickPrice{},
		"ccy",
		[]QuickAccumBook{
			QuickAccumBook{"2010-10-01", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy", -300, []string{"Expense:C"}},
				QuickAccumPosting{"Income", "ccy", 300, []string{"Income"}},
				QuickAccumPosting{"Income:C", "ccy", 100, []string{"Income", "C"}},
				QuickAccumPosting{"Income:T1", "ccy", 200, []string{"Income", "T1"}},
				QuickAccumPosting{"Income:T1:A", "ccy", 100, []string{"Income", "T1", "A"}},
				QuickAccumPosting{"Income:T1:B", "ccy", 100, []string{"Income", "T1", "B"}},
			}},
			QuickAccumBook{"2010-10-02", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy", -300, []string{"Expense:C"}},
				QuickAccumPosting{"Income", "ccy", 300, []string{"Income"}},
				QuickAccumPosting{"Income:C", "ccy", 100, []string{"Income", "C"}},
				QuickAccumPosting{"Income:T1", "ccy", 200, []string{"Income", "T1"}},
				QuickAccumPosting{"Income:T1:A", "ccy", 100, []string{"Income", "T1", "A"}},
				QuickAccumPosting{"Income:T1:B", "ccy", 100, []string{"Income", "T1", "B"}},
			}},
			QuickAccumBook{"2010-10-02", "rand2", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy", -300, []string{"Expense:C"}},
				QuickAccumPosting{"Income", "ccy", 300, []string{"Income"}},
				QuickAccumPosting{"Income:C", "ccy", 100, []string{"Income", "C"}},
				QuickAccumPosting{"Income:T1", "ccy", 200, []string{"Income", "T1"}},
				QuickAccumPosting{"Income:T1:A", "ccy", 100, []string{"Income", "T1", "A"}},
				QuickAccumPosting{"Income:T1:B", "ccy", 100, []string{"Income", "T1", "B"}},
			}},
		},
	)
}

func TestAccumCCY(t *testing.T) {
	// Test different currencies for the same account.
	ValidBook(t,
		[]QuickBook{
			QuickBook{"2010-10-01", "rand", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy1", 100},
				QuickPosting{"Income:T1:A", "ccy2", 100},
				QuickPosting{"Expense:C", "ccy1", -100},
				QuickPosting{"Expense:C", "ccy2", -100},
			}},
		},
		[]QuickPrice{
			QuickPrice{"ccy1", "ccy", 1},
			QuickPrice{"ccy2", "ccy", 2},
		},
		"ccy",
		[]QuickAccumBook{
			QuickAccumBook{"2010-10-01", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy", -300, []string{"Expense:C"}},
				QuickAccumPosting{"Expense:C", "ccy1", -100, []string{"Expense:C", ""}},
				QuickAccumPosting{"Expense:C", "ccy2", -100, []string{"Expense:C", ""}},
				QuickAccumPosting{"Income:T1:A", "ccy", 300, []string{"Income:T1:A"}},
				QuickAccumPosting{"Income:T1:A", "ccy1", 100, []string{"Income:T1:A", ""}},
				QuickAccumPosting{"Income:T1:A", "ccy2", 100, []string{"Income:T1:A", ""}},
			}},
		},
	)
}

func TestAccumCCYSimple(t *testing.T) {
	// Test different single currency
	// TODO: Level isn't right here. Or is it?
	ValidBook(t,
		[]QuickBook{
			QuickBook{"2010-10-01", "rand", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy2", 100},
				QuickPosting{"Expense:C", "ccy2", -100},
			}},
		},
		[]QuickPrice{
			QuickPrice{"ccy2", "ccy1", 2},
		},
		"ccy1",
		[]QuickAccumBook{
			QuickAccumBook{"2010-10-01", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy1", -200, []string{"Expense:C"}},
				QuickAccumPosting{"Expense:C", "ccy2", -100, []string{"Expense:C", ""}},
				QuickAccumPosting{"Income:T1:A", "ccy1", 200, []string{"Income:T1:A"}},
				QuickAccumPosting{"Income:T1:A", "ccy2", 100, []string{"Income:T1:A", ""}},
			}},
		},
	)
}

func TestSimple(t *testing.T) {
	ValidBook(t,
		[]QuickBook{
			QuickBook{"2010-10-01", "rand", []QuickPosting{
				QuickPosting{"Equity", "ccy1", -50},
				QuickPosting{"Mortgage", "ccy1", -100},
				QuickPosting{"Savings:A", "ccy1", 50},
				QuickPosting{"Savings:B", "ccy1", 50},
				QuickPosting{"Savings:C", "ccy1", 50},
			}},
		},
		[]QuickPrice{},
		"ccy1",
		[]QuickAccumBook{
			QuickAccumBook{"2010-10-01", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Equity", "ccy1", -50, []string{"Equity"}},
				QuickAccumPosting{"Mortgage", "ccy1", -100, []string{"Mortgage"}},
				QuickAccumPosting{"Savings", "ccy1", 150, []string{"Savings"}},
				QuickAccumPosting{"Savings:A", "ccy1", 50, []string{"Savings", "A"}},
				QuickAccumPosting{"Savings:B", "ccy1", 50, []string{"Savings", "B"}},
				QuickAccumPosting{"Savings:C", "ccy1", 50, []string{"Savings", "C"}},
			}},
		},
	)

}

func TestTermNonTermMix(t *testing.T) {
	ValidBook(t,
		[]QuickBook{
			QuickBook{"2010-10-01", "rand", []QuickPosting{
				QuickPosting{"Income:T1:A", "ccy1", 100},
				QuickPosting{"Income:T1", "ccy1", 100},
				QuickPosting{"Expense:C", "ccy1", -200},
			}},
		},
		[]QuickPrice{
			QuickPrice{"ccy1", "ccy2", 2},
		},
		"ccy1",
		[]QuickAccumBook{
			QuickAccumBook{"2010-10-01", "rand", []QuickAccumPosting{
				QuickAccumPosting{"Expense:C", "ccy1", -200, []string{"Expense:C"}},
				QuickAccumPosting{"Income:T1", "ccy1", 200, []string{"Income:T1"}},
				QuickAccumPosting{"Income:T1", "ccy1", 100, []string{"Income:T1", ""}},
				QuickAccumPosting{"Income:T1:A", "ccy1", 100, []string{"Income:T1", "A"}},
			}},
		},
	)
}
