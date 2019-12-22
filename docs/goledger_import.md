## goledger import

Import transactions

### Synopsis

Import transactions

```
goledger import [flags]
```

### Options

```
  -a, --acct string                 account for imported postings
  -c, --cacct string                counteraccount for new transactions
      --code string                 code for import or file:<file> for external code (see help code)
  -d, --dedup                       deduplicate transactions based on payee and date (default true)
      --format <type>[:<k=v,...>]   format of input (see help format)
  -h, --help                        help for import
  -r, --reclassify                  reclassify the counteraccount based on previous transactions (default true)
```

### Options inherited from parent commands

```
      --all              all accounts, not just non-zero balance
      --ccy string       base currency
      --colour           colour (ansi) for reports (default true)
      --divider string   divider for account components for reports (default ":")
      --lang string      language
  -l, --ledger string    ledger to read (default "main.ledger")
      --verbose          verbose
```

### SEE ALSO

* [goledger](goledger.md)	 - goledger text-based account application
* [goledger import bankformat](goledger_import_bankformat.md)	 - Bank Format

