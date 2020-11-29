package reports

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"regexp"
	"strconv"
	"strings"
)

var BookOperationUsage = `Book Operations

Note:
  For regular expressions syntax used is used is Golang's.
  This is available here:
    https://github.com/google/re2/wiki/Syntax.

Operations:
  map=/search-regex/replace-regex/(alternate-account/)?

    All accounts that match search-regex will have the replace-regex
    substituted. Captured groups can be replaced using $1, $2, etc.

	The alternate account is used if, if specified and non-empty,
	for the account to substitute if the search doesn't match.

    Example:
    map=/^Asset:[^:]*:(.*)$/Asset:$1/

    Convert 3+ level Asset accounts into 2+ levels by stripping the
    2nd level.

  move=/search-regex/new-account/factor/

    All accounts matching search-regex -- for their postings -- will
    have a new pair of postings applied to transfer the posting amount
    multpled by factor into the new-account. The new-account can used
    captured groups from the search-regex.

    Example:
    move=/^Expense:Regular:([^:]+)$/Expense:Irregular:$1/0.2/

    Re-direct (through new transfer posting) 0.2 of all regular expenses
    into the Expense:Irregular category.

  asof=date
  since=date

    This sets the asof date of the book (all postings up to and excluding
    asof date), and the since date (all postings since and including the
    since date).

    Date can be of formart "YYYY-MM-DD" or "(this|next|last) (year|month|quarter)".

    This is the closing year/month/quarter from today, next/last is the one
    after and the one before.

    Example:
    asof="this year"
    since="last year"

    This will include everything since the preceeding Jan 1st (including Jan 1st),
    up to the subsequent Jan 1st (excluding Jan 1st).

  combine=type

    Type can be yearly, quarterly, monthly, or all. This will floor all
    transaction dates according to the rule.

  depreciate=/search-regex/asset-acccount/months/

    For all matching accounts, immediately transfer the transaction into asset-account,
    and then over the specified months transfer it back into the matching account a
    straight-line portion of it.

`

// Operation must match this regular expression prefix
var re_op_prefix = regexp.MustCompile("^[A-Za-z][0-9A-Za-z_]*=")

// Regexp for map and move
var map_op = regexp.MustCompile("^/([^/]+)/([^/]+)/(([^/]+)/)?$")
var move_op = regexp.MustCompile("^/([^/]+)/([^/]+)/([0-9\\.]+)/$")
var deprec_op = regexp.MustCompile("^/([^/]+)/([^/]+)/([0-9\\.]+)/$")

func BookOp(op string, b *book.Book, macros map[string][]string) error {
	var err error

	m, ok := macros[op]
	if ok {
		for _, macro := range m {
			err = BookOp(macro, b, macros)
			if err != nil {
				return fmt.Errorf("macro expansion '%s': %v", op, err)
			}
		}
		return nil
	}

	op_type := re_op_prefix.FindString(op)
	if op_type == "" {
		return fmt.Errorf("missing type: must be format '%s' or a valid macro", re_op_prefix.String())
	}

	op_act := op[len(op_type):]

	switch op_type {
	case "map=":
		args := map_op.FindStringSubmatch(op_act)
		if args == nil {
			return fmt.Errorf("map operation '%s', invalid: must be format '%s'", op_act, map_op.String())
		}
		b.RegexAccounts(args[1], args[2], args[4])
		return nil
	case "move=":
		args := move_op.FindStringSubmatch(op_act)
		if args == nil {
			return fmt.Errorf("move operation '%s', invalid: must be format '%s'", op_act, move_op.String())
		}
		factor, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			return fmt.Errorf("move factor '%s', invalid: must be float: %v", args[3], err)
		}
		b.AdjustPost(args[1], args[2], factor)
		return nil
	case "asof=":
		d := book.DateFromString(op_act)
		if d == book.Date(0) {
			return fmt.Errorf("asof date '%s', invalid", op_act)
		}
		b.FilterByDateAsof(d)
		return nil
	case "since=":
		d := book.DateFromString(op_act)
		if d == book.Date(0) {
			return fmt.Errorf("since date '%s', invalid", op_act)
		}
		b.FilterByDateSince(d)
		return nil
	case "combine=":
		for _, typ := range book.FloorTypes {
			if strings.EqualFold(op_act, typ) {
				b.SplitBy(typ)
				return nil
			}
		}
		return fmt.Errorf("combine type '%s', invalid: must be one of %s",
			op_act, strings.Join(book.FloorTypes, ","))
	case "depreciate=":
		args := deprec_op.FindStringSubmatch(op_act)
		if args == nil {
			return fmt.Errorf("depreciate operation '%s', invalid: must be format '%s'", op_act, deprec_op.String())
		}
		months, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("depreciate months '%s', invalid: must be integer: %v", args[3], err)
		}
		b.Depreciate(args[1], args[2], "monthly", months)
		return nil
	default:
		return fmt.Errorf("operation type '%s' invalid: must be one of map, move, since, asof, combine, or depreciate", op_type)
	}
}
