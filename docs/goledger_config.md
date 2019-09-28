Configuration File

The configuration file mirrors largely mirrors parameters available
on the command line.

There are two notable exceptions:
  - report.macros
    This section defines macros for the report (set of operations) that
    can be used on the command line arguments or by other macros. This
    is how complex reports can be built up.

  - importdefs.<name>
    Defining a new importdefs <name> will create a sub-command under
    import that has all of the command line parameters and the import
    configuration pre-configured. This allows you to create an import
    definition per CSV file (or other format) that you download.

  - register.accounts
    List of accounts that shell-completion will match if used. This is
    to make it easier to use the CLI.

Starter configuration file:
```

#
# Main defaults
#
#ledger =  "default_ledger_file"
#baseccy = "£"

#
# Defaults for the report command
#
[report]
<<<<<<< HEAD
combineby = "today"
type =      "Ansi"
sum =       true
convert =   true
credit =    "^(Income|Trading|Liability|Equity)(:.*)?$"

[report.macros]
macroA = [
	"book operation A",
	"book operation B",
]

#
# Defaults for the register command
#
<<<<<<< HEAD

[register]
accounts = [
  "Asset:Default",
]
count = -100
asc = true

[importdefs.bankformat]
description = "Bank Format"
configtype = "csv"
account = "Asset:BankDefaultAccount"
counteraccount = "Expense:DefaultExpenseAccount"
dedup = true
reclassify = true

[importdefs.bankformat.params]
ccy = "£"
date = "0"
amount = "2"
payee = "1"

```

