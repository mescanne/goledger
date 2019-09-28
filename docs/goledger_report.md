## goledger report

Aggregated transaction reports

### Synopsis

Aggregated transactions reports

There are two basic dimensions for transactions reports:

Time period:
  If you look at a time period from the beginning of time
  until a point in time (eg now) you will see the total
  balance of all of the accounts.
  
  If you select a start and end date (eg beginning of this
  year until now) you will see the change in balance
  across all of the accounts.

Account Regexp:
  Normally, you don't want to see all accounts but focus
  on a particular subset of the accounts. Or a certain
  categorisation of accounts.
  
  Using regular expressions you can create income statements,
  balance sheets, and cashflow statements.

  Example for a balance sheet:
  Map all ^Income:.* and ^Expense.:* accounts into Equity. Also
  include all other accounts that aren't Asset:, Liability:,
  or Equity.

  This will leave just Asset and Liabilities.
    


```
goledger report [macros|ops...] [flags]
```

### Options

```
      --convert             convert to base currency
      --credit string       credit account regex
  -h, --help                help for report
      --splitby floorType   combine transactions by periodic date (values none, yearly, quarterly, monthly, today)
      --sum                 summarise transactions
      --type reportType     report type (Text, Ledger)
```

### Options inherited from parent commands

```
      --ccy string       base currency
      --colour           colour (ansi) for reports (default true)
      --divider string   divider for account components for reports (default ":")
      --lang string      language (default "en_GB.UTF-8")
  -l, --ledger string    ledger to read (default "main.ledger")
      --verbose          verbose
```

### SEE ALSO

* [goledger](goledger.md)	 - goledger text-based account application

