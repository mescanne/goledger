package script

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

func Print(_ *starlark.Thread, msg string) {
	fmt.Printf("; %s\n", msg)
}

var userError = errors.New("user code")
var Errorf = starlark.NewBuiltin("error",
	func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var msg string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "msg", &msg); err != nil {
			return nil, err
		}
		return starlark.None, fmt.Errorf("%w: %s", userError, msg)
	})

func GetBuilderFunction(bbuilder *book.Builder, account string, ccy string, counterAccount string) starlark.Value {
	buildf := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

		// Mandatory values
		var date string
		var desc string
		var amount starlark.Value

		// Optional first account values
		var denom int = 1
		var ccy string = ccy
		var account string = account

		// Optional second account
		var amount2 starlark.Value = starlark.None
		var ccy2 string = ccy
		var denom2 int = 1
		var account2 string = ""

		// Optional final settling account
		var caccount string = counterAccount
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
		amt, err := GetBigRat(amount)
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
		amt2, err := GetBigRat(amount2)
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
	return starlark.NewBuiltin("add", buildf)
}

func RunStarlark(printf func(*starlark.Thread, string), globals map[string]starlark.Value, code string) error {
	sc, err := utils.GetFileOrStr(code)
	if err != nil {
		return err
	}

	// Make more flexible (esp floats)
	resolve.AllowFloat = true
	resolve.AllowGlobalReassign = true
	resolve.AllowRecursion = true

	// The Thread defines the behavior of the built-in 'print' function.
	thread := &starlark.Thread{
		Name:  "main",
		Print: printf,
	}

	// This dictionary defines the pre-declared environment.
	predeclared := starlark.StringDict(globals)

	// Execute a program.
	_, err = starlark.ExecFile(thread, "code", sc, predeclared)
	if err != nil {
		if errors.Is(err, userError) {
			return err
		}

		// Full backtrace if possible
		if evalErr, ok := err.(*starlark.EvalError); ok {
			fmt.Fprintf(os.Stderr, "Runtime error stack: %s\n", evalErr.Backtrace())
		}

		return fmt.Errorf("running: %w", err)
	}

	return nil
}
