## goledger import

Import transactions

### Synopsis

Import transactions

```
goledger import [flags]
```

### Options

```
  -a, --acct string               Account for imported postings
  -c, --cacct string              Counteraccount for new transactions
  -d, --dedup                     Deduplicate transactions based on payee and date (default true)
      --format <type>:<k=v;...>   Format of input (see help format)
  -h, --help                      help for import
  -r, --reclassify                Reclassify the counteraccount based on previous transactions (default true)
```

### Options inherited from parent commands

```
      --ccy string       Base Currency
      --colour           Colour (ansi) for reports (default true)
      --divider string   Divider for account components for reports (default ":")
      --lang string      Language (default "en_GB.UTF-8")
  -l, --ledger string    Ledger to read (default "main.ledger")
      --verbose          Verbose
```

### SEE ALSO

* [goledger](goledger.md)	 - goledger text-based account application

