
For a given account A:
  For BaseCCY:
    - Any transfers in and out are cashflows +/-
    - Any income/expense should simply be excluded. Income/expense will reflect in a change in - in the filter.
  For investment CCY:
    - Any trades from baseCCY into CCY is +
    - Any trades out is -

Basic info for a portfolio:
 - Opening balance for a period.
 - Closing balance for a period.
 - Net inflow, outflow
 - Gains/losses.
 - Percentage return.

We can have ->
 - Multple portfolios. And combine those portfolios.
 - Multiple periods of time. (annual, quarterly, for the past N years) And combine those periods of time.

X axis periods of time and Y axis portfolios.

For each timeperiod and portfolio:
  - % return
  - net inflow/outflow
  - gains/losses

At the end:
  - closing balance

At the beginning:
  - opening balance

                         POrtA              PortB        PortC       All
			 BAl   In/Out  IRR
Opening Balance          10                 10      10      30
Date snapshot            12
Date snapshot            13
Closing Balance (date)   13

Cons:
 - Opening balance only applies to balance, not in/out, IRR, or gains
 - Closing balance date is a partial period. That's fine.
 - No clear place for IRR aggregated across entire period.


                         Bal                       In/Out                   IRR
			 PortA  PortB  PortC  All  PortA   PortB  PortC All PortA  PortB
Opening Balance          10                 10      10      30
Date snapshot            12
Date snapshot            13
Closing Balance (date)   13

Cons:
 - Even more murky aobut in/out and IRR not being there for opening balance


Question:
 - Is it really important to combine in one report multiple portfolios (medium)
 - Is it important to combine in one report multiple periods (high)
 - Is it important to provide a detail view of activity (yes - optional, perhaps)

Given that, what we'd like in terms of detail:
Options: All rates are annualised or total?
 - Opening date
 - 

                        Open  DateA   DateB  Close
PortA Bal
      In/Out
	Gains
	IRR
PortB
	Bal
	In/Out
	Gains
	IRR
Combined
	Bal
	In/Out
	GAins
	IRR

Downsides:
 - Unlimited number of columns to the left. No good.
