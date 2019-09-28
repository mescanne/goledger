## goledger register

Show registry of account postings

### Synopsis

Register account postings

Show a registry of postings for an individual account. This
is useful for reconciliation between accounts and for investigating
one account.


```
goledger register [acct regex]... [flags]
```

### Options

```
      --asc            ascending or descending order (default true)
      --asof string    end date
      --begin string   begin date
      --count int      count of entries (0 = no limit) (default -100)
  -h, --help           help for register
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

