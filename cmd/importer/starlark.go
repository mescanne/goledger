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

func (imp *ImportDef) processData(idata starlark.Value, sc string) (*book.Book, error) {

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

		// Optional extra values
		var denom int = 1
		var ccy string = imp.CCY
		var note string = ""
		var lnote string = ""
		var account string = imp.Account
		var caccount string = imp.CounterAccount

		err := starlark.UnpackArgs(b.Name(), args, kwargs,
			"date", &date, "desc", &desc, "amt", &amount,
			"ccy?", &ccy, "denom?", &denom,
			"account?", &account, "caccount?", &caccount,
			"note?", &note, "lnote?", &lnote)
		if err != nil {
			return nil, err
		}

		// Convert the date
		ndate := book.DateFromString(string(date))
		if ndate == book.Date(0) {
			return nil, fmt.Errorf("invalid date %s", date)
		}

		// Set the amount
		var amt *big.Rat = &big.Rat{}
		switch v := amount.(type) {
		case starlark.Float:
			amt.SetFloat64(float64(v))
		case starlark.Int:
			amt.SetInt(v.BigInt())
		case starlark.String:
			_, ok := amt.SetString(string(v))
			if !ok {
				return nil, fmt.Errorf("invalid amount '%s'", v)
			}
		default:
			return nil, fmt.Errorf("not a valid amount type %T (string, int, float)", amount)
		}

		// Divide by denominator if set
		amt.Mul(amt, big.NewRat(1, int64(denom)))

		// Set the negative of the amount
		neg := (&big.Rat{}).Neg(amt)

		// Make the transaction
		bbuilder.NewTransaction(ndate, string(desc), note)
		bbuilder.AddPosting(account, ccy, amt, lnote)
		bbuilder.AddPosting(caccount, ccy, neg, "")
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
