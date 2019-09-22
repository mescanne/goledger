Preconfigured macros
  
  Macro bal
    map=/^(Income|Expense|Equity|Trading).*$/Equity/
    map=/^(Asset):([^:]*):(Pension):(.*)$/$1:$3:$2:$4/
    map=/^(Asset|Liability):([^:]*):([^P][^e][^:]*):/$1:$3:$2/
  
  Macro broadstairs-exp
    map=/^(.*)$/Other:$1/
    map=/^Other:Income:Parents(:.*)?$/Funds:Parents/
    map=/^Other:Expense:SharedBroadstairs:(.*)$/Expense:$1/
    map=/^Other(:.*)?$/Funds:Joint/
  
  Macro broadstairs-fund
    map=/^Expense:(.*)$/Joint:Expense/
    map=/^Funds:(.*)$/$1:Funds/
  
  Macro cf
    map=/^(Income|Expense|Trading|Equity)(:.*)?$/Cashflow/
  
  Macro compact
    map=/^(Expense):([^:]*)(:.*)$/$1:$2/
  
  Macro is
    map=/^(Asset|Liability|Trading|Equity)(:.*)?$/Operating Income/
  
  Macro normal
    map=/^Income:([^:]*):([^:]*(Bonus|GSU|Options))(:.*)?$/Income:Bonus:$2$4/
    map=/^Income:([^:]*):([^:]*Salary)(:.*)?$/Income:Salary:$2$3/
    map=/^Income:([^:]*):([^:]*Pension)(:.*)?$/Income:Pension:$2$3/
    map=/^Income:(Mark|Patty):(.*)$/Income:Other:$2/
    map=/^Expense:Mark:Computing/Expense:Mark-Computing/
    map=/^Expense:(Mark|Patty|Joint):(.*)$/Expense:$2/
    map=/^Expense:(.*)$/Expense:Regular:$1/
    map=/^Expense:Regular:Childcare:(.*)$/Expense:Childcare:$1/
    map=/^Expense:Regular:Cash$/Expense:Cash/
    map=/^Expense:Regular:School:(.*)$/Expense:School:$1/
    map=/^Expense:Regular:SharedBroadstairs:(.*)$/Expense:SharedBroadstairs:$1/
    map=/^Expense:Regular:Travel:(.*)$/Expense:Travel:$1/
    map=/^Asset:([^:]*):([^:]*):(.*)$/Asset:$2:$1 $3/
  
  Macro offset
    map=/^(.*)$/Other:$1/
    map=/^Other:Asset:(.*):Current:(FirstDirect.*)$/Offset Accounts:$1 $2/
    map=/^Other:.*FDMortgage$/Offset Mortgage/
    map=/^Other:.*$/Equity/
  
  Macro thismonth
    since=this quarter
  
  Macro thisquarter
    since=this quarter
  
  Macro thisyear
    since=this year
  


### SEE ALSO

* [goledger report](goledger_report.md)	 - Aggregated transaction reports

