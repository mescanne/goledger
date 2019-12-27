package importer

import (
	"errors"
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/utils"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"math/big"
	"os"
)

func getNumber(amount starlark.Value) (*big.Rat, error) {
	if amount == starlark.None {
		return nil, nil
	}

	var amt *big.Rat = &big.Rat{}
	switch v := amount.(type) {
	case starlark.Float:
		amt.SetFloat64(float64(v))
	case starlark.Int:
		amt.SetInt(v.BigInt())
	case starlark.String:
		_, ok := amt.SetString(string(v))
		if !ok {
			return nil, fmt.Errorf("invalid amount '%s'", string(v))
		}
	default:
		return nil, fmt.Errorf("not a valid amount type %T (string, int, float)", amount)
	}

	return amt, nil
}

func (imp *ImportDef) processData(idata starlark.Value, file string, sc string) (*book.Book, error) {

	sc, err := utils.GetFileOrStr(sc)
	if err != nil {
		return nil, err
	}

	// Make more flexible (esp floats)
	resolve.AllowFloat = true
	resolve.AllowGlobalReassign = true
	resolve.AllowRecursion = true

	// Capture user errors
	userError := errors.New("user code")
	errorf := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var msg string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "msg", &msg); err != nil {
			return nil, err
		}
		return starlark.None, fmt.Errorf("%w: %s", userError, msg)
	}

	// Build the book
	bbuilder := book.NewBookBuilder()
	addf := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		// Mandatory values
		var date string
		var desc string
		var amount starlark.Value

		// Optional first account values
		var denom int = 1
		var ccy string = imp.CCY
		var account string = imp.Account

		// Optional second account
		var amount2 starlark.Value = starlark.None
		var ccy2 string = imp.CCY
		var denom2 int = 1
		var account2 string = ""

		// Optional final settling account
		var caccount string = imp.CounterAccount
		var note string = ""
		var lnote string = ""

		err := starlark.UnpackArgs(b.Name(), args, kwargs,
			"date", &date, "desc", &desc,
			"amt", &amount, "ccy?", &ccy, "denom?", &denom, "account?", &account,
			"amt2", &amount2, "ccy2?", &ccy2, "denom2?", &denom2, "account2?", &account2,
			"caccount?", &caccount,
			"note?", &note,
			"lnote?", &lnote)
		if err != nil {
			return nil, err
		}

		// Convert the date
		ndate := book.DateFromString(string(date))
		if ndate == book.Date(0) {
			return nil, fmt.Errorf("invalid date %s", date)
		}

		// Convert the amount
		amt, err := getNumber(amount)
		if err != nil {
			return nil, err
		}

		// Divide by denominator if set
		amt.Mul(amt, big.NewRat(1, int64(denom)))

		// Set the negative of the amount
		neg := (&big.Rat{}).Neg(amt)

		// Make the transaction
		bbuilder.NewTransaction(ndate, string(desc), note)
		bbuilder.AddPosting(account, ccy, amt, lnote)
		bbuilder.AddPosting(caccount, ccy, neg, "")

		// Nothing else to do
		if amount2 == starlark.None && account2 == "" {
			return starlark.None, nil
		}

		if amount2 == starlark.None || account2 == "" || ccy2 == ccy {
			// Error
		}

		// Convert the amount2
		amt2, err := getNumber(amount2)
		if err != nil {
			return nil, err
		}

		// Divide by denominator if set
		amt2.Mul(amt2, big.NewRat(1, int64(denom2)))

		// Set the negative of the amount
		neg2 := (&big.Rat{}).Neg(amt2)

		bbuilder.AddPosting(account2, ccy2, amt2, "")
		bbuilder.AddPosting(caccount, ccy2, neg2, "")

		return starlark.None, nil
	}

	// The Thread defines the behavior of the built-in 'print' function.
	thread := &starlark.Thread{
		Name: "main",
		Print: func(_ *starlark.Thread, msg string) {
			fmt.Printf("; %s\n", msg)
		},
	}

	// This dictionary defines the pre-declared environment.
	predeclared := starlark.StringDict{
		"data":  idata,
		"file":  starlark.String(file),
		"add":   starlark.NewBuiltin("add", addf),
		"error": starlark.NewBuiltin("error", errorf),
	}

	// Execute a program.
	_, err = starlark.ExecFile(thread, "code", sc, predeclared)
	if err != nil {
		if errors.Is(err, userError) {
			return nil, err
		}

		// Full backtrace if possible
		if evalErr, ok := err.(*starlark.EvalError); ok {
			fmt.Fprintf(os.Stderr, "Runtime error stack: %s\n", evalErr.Backtrace())
		}

		return nil, fmt.Errorf("importing: %w", err)
	}

	// Build the book -- done!
	return bbuilder.Build(), nil
}
