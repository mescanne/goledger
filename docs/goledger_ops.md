Book Operations

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

    Type can be yearly, quarterly, monthly, or today. This will floor all
    transaction dates according to the rule.


