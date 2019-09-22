
Blah blah
```

#
# Main defaults
#
## Default ledger file
#ledger =  "default_ledger_file"
## Default base CCY (for reporting)
#baseccy = "£"

#
# Defaults for the report command
#
[report]
# combineby = "today"
# type =      "Ansi"
# sum =       true
# convert =   true
# credit =    "^(Income|Trading|Liability|Equity)(:.*)?$"
#
#[transactioncmd.macros]
#macroA = [
#	"book operation A",
#	"book operation B",
#]

#
# Defaults for the register command
#
[register]
# accounts = [
#   "Asset:<Default>",
# ]
# count = -100
# asc = true

# Repeated import definitions
# for each custom-importer
[[importdefs]]
#name = "fdcurr"
#description = "FirstDirect Current Account"
#configtype = "csv"
#account = "Asset:Joint:Current:FirstDirectCurrent"
#counteraccount = "Expense:Joint:Miscellaneous"
#dedup = true
#reclassify = true
#params = { ccy = "£", date = "0", amount = "2", payee = "1" }
```
Blah blah


