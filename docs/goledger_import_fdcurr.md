## goledger import fdcurr

FirstDirect Current Account

### Synopsis

Import transactions

```
goledger import fdcurr [flags]
```

### Options

```
  -a, --acct string               Account for imported postings (default "Asset:Joint:Current:FirstDirectCurrent")
  -c, --cacct string              Counteraccount for new transactions (default "Expense:Joint:Miscellaneous")
  -d, --dedup                     Deduplicate transactions based on payee and date (default true)
      --format <type>:<k=v;...>   Format of input (see help format) (default csv:payee=1;ccy=%C2%A3;date=0;amount=2)
  -h, --help                      help for fdcurr
  -r, --reclassify                Reclassify the counteraccount based on previous transactions (default true)
```

### Options inherited from parent commands

```
      --ccy string       Base Currency (default "£")
      --colour           Colour (ansi) for reports (default true)
      --divider string   Divider for account components for reports (default ":")
      --lang string      Language (default "en_GB.UTF-8")
  -l, --ledger string    Ledger to read (default "/home/mescanne/docs/ledger/main.ledger")
      --verbose          Verbose
```

### SEE ALSO

* [goledger import](goledger_import.md)	 - Import transactions

