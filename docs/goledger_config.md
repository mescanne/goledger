
Blah blah
```

#
# Main defaults
#
# ledger =  "default_ledger_file"
# baseccy = "£"

#
# Defaults for the transaction command
#
[transactioncmd]
# combineby = "today"
# type =      "Ansi"
# sum =       true
# convert =   true
# credit =    "^(Income|Trading|Liability|Equity)(:.*)?$"

#
# Defaults for the register command
#
[registercmd]
accounts = [
  "Asset:<Default>",
]
# count = -100
# asc = true

[[importdefs]]
name = "fdcurr"
description = "FirstDirect Current Account"
configtype = "csv"
account = "Asset:Joint:Current:FirstDirectCurrent"
counteraccount = "Expense:Joint:Miscellaneous"
dedup = true
reclassify = true
[importdefs.params]
ccy = "£"
date = "0"
amount = "2"
payee = "1"

[macros]
macroA = [
	"book operation A",
	"book operation B",
]
```
Blah blah


### SEE ALSO

* [goledger](goledger.md)	 - goledger text-based account application

