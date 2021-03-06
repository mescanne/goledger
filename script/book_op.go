package script

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"go.starlark.net/starlark"
)

var mapf = starlark.NewBuiltin("map",
	func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var from, to, alt string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs,
			"from", &from,
			"to", &to,
			"alt?", &alt); err != nil {
			return nil, err
		}
		val := b.Receiver()
		if val == nil {
			return starlark.None, fmt.Errorf("%w: %s", userError, "unbound method")
		}
		bb, ok := val.(*starlarkBook)
		if !ok {
			return starlark.None, fmt.Errorf("%w: %s", userError, "bound to not a book")
		}
		bb.b.RegexAccounts(from, to, alt)
		return val, nil
	})

var asoff = starlark.NewBuiltin("asof",
	func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var date string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "date", &date); err != nil {
			return nil, err
		}
		val := b.Receiver()
		if val == nil {
			return starlark.None, fmt.Errorf("%w: %s", userError, "unbound method")
		}
		bb, ok := val.(*starlarkBook)
		if !ok {
			return starlark.None, fmt.Errorf("%w: %s", userError, "bound to not a book")
		}
		dstr := book.DateFromString(date)
		if dstr == book.Date(0) {
			return starlark.None, fmt.Errorf("invalid date: '%s'", date)
		}
		bb.b.FilterByDateAsof(dstr)
		return val, nil
	})

var sincef = starlark.NewBuiltin("since",
	func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var date string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "date", &date); err != nil {
			return nil, err
		}
		val := b.Receiver()
		if val == nil {
			return starlark.None, fmt.Errorf("%w: %s", userError, "unbound method")
		}
		bb, ok := val.(*starlarkBook)
		if !ok {
			return starlark.None, fmt.Errorf("%w: %s", userError, "bound to not a book")
		}
		dstr := book.DateFromString(date)
		if dstr == book.Date(0) {
			return starlark.None, fmt.Errorf("invalid date: '%s'", date)
		}
		bb.b.FilterByDateSince(dstr)
		return val, nil
	})

type starlarkBook struct {
	b      *book.Book
	frozen bool
}

func (s *starlarkBook) Attr(name string) (starlark.Value, error) {

	// Read-only methods
	switch name {
	case "transactions":
		return starlarkTransactions(s.b.Transactions()), nil
	}

	if s.frozen {
		return nil, fmt.Errorf("operation %s invalid: book is frozen")
	}

	// Modifying methods
	switch name {
	case "since":
		return sincef.BindReceiver(s), nil
	case "asof":
		return asoff.BindReceiver(s), nil
	case "map":
		return mapf.BindReceiver(s), nil
	default:
		return nil, nil
	}
}
func (s *starlarkBook) AttrNames() []string {
	return []string{
		"transactions",
		"since",
		"asof",
		"map",
	}
}
func (s *starlarkBook) String() string {
	return "<n/a>book"
}
func (s *starlarkBook) Type() string {
	return "book"
}
func (s *starlarkBook) Freeze() { s.frozen = true }
func (s *starlarkBook) Truth() starlark.Bool {
	return true
}
func (s *starlarkBook) Hash() (uint32, error) {
	return 0, fmt.Errorf("cannot hash book")
}

func ConvertBookToStarlark(main *book.Book) starlark.Value {
	return &starlarkBook{
		b:      main,
		frozen: false,
	}
}
