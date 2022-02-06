package trading

import (
	"fmt"
	"github.com/mescanne/goledger/book"
	"github.com/mescanne/goledger/cmd/app"
	//"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"math"
	"math/big"
	"sort"
)

// Configuration for a Register Report
type TradingReport struct {
	Name           string
	Short          string
	Long           string
	EndDate        string
	CCY            string
	TradingAccount string
}

const trading_long = `Register account postings

Show a registry of postings individual accounts. This is useful for reconciliation
between accounts and for investigating postings.
`

/*
 * Trading account:
 * This represents things coming and going from your system of accounts. Traded away. Or traded in.
 * You *may* be able to infer a rate from it. You may not.
 *
 * More precisely, for sources of rates you have
 *   - prices database
 *   - trades in and out of your asset account
 *
 * Analysis most accurate is -
 *  Focus on (account, symbol)
 */

func Add(cmd *cobra.Command, app *app.App, rep *TradingReport) {
	ncmd := &cobra.Command{
		Args: cobra.MinimumNArgs(1),
		// Set of CCYs
		//ValidArgs:         reg.Accounts,
		Use:               "trading <ccy...>",
		Long:              trading_long,
		Short:             "Show trading report",
		DisableAutoGenTag: true,
	}

	// Set defaults
	//ncmd.Flags().Var(reportType, "type", fmt.Sprintf("report type (%s)", reportType.Values()))
	ncmd.Flags().StringVar(&rep.EndDate, "asof", rep.EndDate, "end date")
	ncmd.Flags().StringVar(&rep.TradingAccount, "account", rep.TradingAccount, "trading account")
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return rep.run(app, cmd, args)
	}

	cmd.AddCommand(ncmd)
}

func (rep *TradingReport) run(rapp *app.App, cmd *cobra.Command, args []string) error {

	// Load up saved flags
	b, err := rapp.LoadBook()
	if err != nil {
		return err
	}

	// Filter by time
	b.FilterByDateAsof(book.DateFromString(rep.EndDate))

	trades, err := process2(b, rapp.BaseCCY, rep.TradingAccount)
	if err != nil {
		return err
	}

	bp := rapp.NewBookPrinter(b.GetCCYDecimals())
	err = ShowText(bp, trades, rapp.BaseCCY)
	if err != nil {
		return err
	}

	/*
	 * Logic:
	 * Search for
	 * - Trading Account.
	 * Expect:
	 * - All Transactions with a Trading account look like:
	 *    --> Four postings. Two CCYs. Trading holds them together.
	 * - Rates (prices) exist for the conversion of the two CCYs.
	 *
	 * Report will -
	 * -
	 */

	//// Apply any operations
	//if len(args) > 1 {
	//	if err = rapp.BookOps(b, args[0:len(args)-1]...); err != nil {
	//		return err
	//	}
	//}

	//// Create printer
	//bp := rapp.NewBookPrinter(b.GetCCYDecimals())

	//// Combined -- just dump out as is
	//arg := args[len(args)-1]
	//if reg.Combined {
	//	re, err := regexp.Compile(arg)
	//	if err != nil {
	//		return fmt.Errorf("invalid regex: '%s': %w", arg, err)
	//	}
	//	return extractRegisterByRegex(b.Transactions(), re, reg.Split).ShowReport(bp, reg.Type, reg.Count, reg.Asc, true, false)
	//}

	//for _, acct := range b.Accounts(arg, !rapp.All) {
	//	bp.Printf("\n%s\n", bp.Ansi(app.BlueUL, acct))
	//	if err := extractRegisterByAccount(b.Transactions(), acct, reg.Split).ShowReport(bp, reg.Type, reg.Count, reg.Asc, false, true); err != nil {
	//		return fmt.Errorf("error writing report '%s': %w", acct, err)
	//	}
	//}

	return nil

}

//
// Approach:
// For a given book and set of CCYs (all?), and baseCCY,  we extract
//  - TradeHistory across all applicable accounts
//  - Map of [ccy] for trade history
//
// Each trade history record has methods to:
//  - dump it out pretty-like.
//  - extract IRR stats for time X. (TradePosition)
//    --> requires just cost, value, and cost-acquisition-date  (easy)
//  - extract trailing IRR for time X.
//    --> requires just cost, value, cost-back-in-time, value-back-in-time, trades since then, etc..
//

type TradeHistory struct {

	// Base CCY
	BaseCCY string

	// Sort key - each transaction
	CCY     string    `json:"ccy"`
	Date    book.Date `json:"date"`
	Account string
	Payee   string `json:"payee,omitempty"`

	// Trade information -- directly from posting
	// Balance is the total for the CCY
	Amount  *big.Rat `json:"amount"`
	Balance *big.Rat `json:"balance"`

	// Inferred from transaction or price book.
	// This is the price paid for the amount (or gained).
	// Always filled in even if inferred/guessed.
	Price       *big.Rat
	PriceSource string

	// Metrics for accumulating trade history.
	Cost        *big.Rat  // Total cost for the balance
	CostAvgDate book.Date // Average date for for purchases
}

// Note:
// We can inject the transaction pries into the pricebook independently. That should be a separate option.
// ... even just extracting implicit prices can be separate.
//
// Operations:
// (with price book)
// - One TradeHistory can calculate a lifetime IRR
// - One TradeHistory (with base currency) could likely calculate a trailing-twelve-month IRR
//
// For each of (BaseCCY, CCY, Account) with the latest date, you can combine all of the
// metrics together.
//  --> add together the cost
//  --> add together cost*date
//
// ... however, you need to be able to add together the market-value. This can be gathered through
// the price book only.
//
// And for non-lifetime, you'd need the cost, and market-value at the previous time.

type TradePosition struct {
	BaseCCY string
	CCY     string
	Date    book.Date

	Amount     *big.Rat // quantity held
	Value      *big.Rat // value in BaseCCY
	Cost       *big.Rat // acquisition value
	CostByDate *big.Rat // acquisition * date value
}

func extractHistory(b *book.Book, acct string, ccy string) ([]*TradeHistory, error) {

	// Only the investment account
	b.FilterByAccount(acct)

	trades := make([]*TradeHistory, 0, 100)

	for _, trans := range b.Transactions() {

		for _, p := range trans {

			if p.GetCCY() == ccy {
				continue
			}

			// Buying/selling in the account
			rate := trans.InferRate(ccy, p.GetCCY())
			rateSource := "Trans"

			// Get rate from pricebook if needed
			if rate == nil {
				r, pt := b.GetPrice(trans.GetDate(), p.GetCCY(), ccy)
				rate = r
				rateSource = pt.String()
			}

			// Calculate
			//avgDaysAgo := int(math.Floor(totalPaidByDate / totalPaid))
			//yearsAgo := float64(trade.Date.AsDays()-avgDaysAgo) / 365.25
			//trades[i].BalValue = big.NewRat(0, 1).SetFloat64(totalBalance * value)

			// This is what's needed.
			// totalPaid must be *prior* to the amount. Or after?
			//trades[i].AvgDate = book.GetDateFromDays(avgDaysAgo)
			// Need prior cost.
			//trades[i].Cost = big.NewRat(0, 1).SetFloat64(totalPaid)

			trades = append(trades, &TradeHistory{
				CCY:  p.GetCCY(),
				Date: trans.GetDate(),

				// Trade detail
				Payee:       trans.GetPayee(),
				Price:       rate,
				PriceSource: rateSource,
				Amount:      p.GetAmount(),

				// To date
				Balance: p.GetBalance(),
			})
		}
	}

	return trades, nil
}

type Trade struct {

	// Key - one row per ccy and date
	CCY  string    `json:"ccy"`
	Date book.Date `json:"date"`

	// For trades - payee, buy/sell amount (quantity) and price (in ref ccy).
	// Optional - empty if only a valuation event
	// If price is empty, assume it is an accumulation event (e.g. no price paid)
	Payee       string `json:"payee,omitempty"`
	Price       *big.Rat
	PriceSource string   // Source of price
	Amount      *big.Rat `json:"amount"`

	// Price (in ref ccy) or Balance
	// Optional - filled in if there's an external value.
	Value       *big.Rat
	ValueSource string // name of source of value price

	// Holdings (inclusive of amount)
	// Always filled in.
	Balance *big.Rat `json:"balance"`

	// Calculated (weighted-average)
	AvgDate  book.Date // Average date for purchase
	Cost     *big.Rat  // Average cost
	BalValue *big.Rat  // Average cost
	IRR      float64
}

//
// IRR needs to be calculated:
// --> upon selling
// --> upon valuation date
//
func isLess(l *Trade, r *Trade) bool {

	// Compare CCY
	if l.CCY < r.CCY {
		return true
	} else if l.CCY > r.CCY {
		return false
	}

	// Compare date
	if l.Date < r.Date {
		return true
	}

	return false
}

// For stats:
//  We have - trades
//          - valuation events
// Plus:    - inferred valuation events (not ideal at all)
//

/* Focus not on trading account, but on investment (brokerage) account.
 *
 * Look for all units.
 *
 * Specify:
 * - Base Unit
 * - Infer trading account?
 *
 */
func process2(b *book.Book, ccy string, acct string) ([]*book.Trade, error) {

	// Only the investment account
	b.FilterByAccount(acct)

	return b.GetTradeHistory(ccy)
}

func process(b *book.Book, ccy string, acct string) ([]*Trade, error) {

	// Only the investment account
	b.FilterByAccount(acct)

	trades := make([]*Trade, 0, 100)

	//zero := big.NewRat(0, 1)

	for _, trans := range b.Transactions() {

		/*
		 * For -
		 *    - each posting matching the investment account with NOT ccy
		 *       --> take this as a purchase or sell (+/-)
		 *    - IF
		 *       --> the counteraccount is singular
		 *       --> the counteraccount's counteraccount is base ccy
		 *       --> infer a rate
		 */

		/*
		 * Or:
		 * Do a map[account][ccy]
		 * And if we found map[account][ccy] len 2, with ccy for found ccy *and* baseCCY,
		 * *and* account is only *other* account for ccy...
		 * Or amount is equal.
		 */

		// make(map[string]map[string]*big.Rat)
		// Or:
		// we know the sorted order.
		// so just check consecutive postings.
		// then infer rates from that.

		for _, p := range trans {

			// Buying/selling in the account
			if p.GetAccount() == acct && p.GetCCY() != ccy {
				rate := trans.InferRate(ccy, p.GetCCY())
				rateSource := "Trans"
				if rate != nil {
					v, _ := rate.Float64()
					fmt.Printf("Inferred rate: %v\n", v)
				} else {
					fmt.Printf("Failed inferring rate\n")
				}
				r, pt := b.GetPrice(trans.GetDate(), p.GetCCY(), ccy)
				rf, _ := r.Float64()
				fmt.Printf("Pricebook: %f, type %s\n", rf, pt)
				if rate == nil {
					rate = r
					rateSource = "Value"
				}
				trades = append(trades, &Trade{
					CCY:  p.GetCCY(),
					Date: trans.GetDate(),

					// Trade detail
					Payee:       trans.GetPayee(),
					Price:       rate,
					PriceSource: rateSource,
					Amount:      p.GetAmount(),

					// To date
					Balance: p.GetBalance(),

					// Extra
					Value:       r,
					ValueSource: pt.String(),
				})
			}
		}
	}

	sort.Slice(trades, func(i, j int) bool {
		return isLess(trades[i], trades[j])
	})

	// Need ->
	//  - Value   (value is needed with total balance for the marketvalue of the portfolio/quantity)
	//  - Amount
	//  - Price
	//  - Date
	var totalPaidByDate, totalPaid, totalBalance float64
	for i, trade := range trades {

		// Reset stats if new CCY or first time
		if i == 0 || trades[i-1].CCY != trade.CCY {
			totalPaid = 0.0
			totalBalance = 0.0
			totalPaidByDate = 0.0
		}

		// Dump stats based on values
		if trade.Value != nil && totalPaid > 0.0 {
			value, _ := trade.Value.Float64()
			avgDaysAgo := int(math.Floor(totalPaidByDate / totalPaid))
			yearsAgo := float64(trade.Date.AsDays()-avgDaysAgo) / 365.25
			trades[i].AvgDate = book.GetDateFromDays(avgDaysAgo)
			trades[i].Cost = big.NewRat(0, 1).SetFloat64(totalPaid)
			trades[i].BalValue = big.NewRat(0, 1).SetFloat64(totalBalance * value)
			trades[i].IRR = 100 * (math.Pow(totalBalance*value/totalPaid, 1/yearsAgo) - 1)
		} else {
			trades[i].AvgDate = book.GetToday()
			trades[i].Cost = big.NewRat(0, 1)
			trades[i].BalValue = big.NewRat(0, 1)
			trades[i].IRR = 0.0
		}

		// Dump stats based on values
		if trade.Amount != nil {
			amount, _ := trade.Amount.Float64()

			// Accumulate stats
			if amount > 0.0 && trade.Price != nil {
				price, _ := trade.Price.Float64()
				paid := amount * price
				totalPaid += paid
				totalPaidByDate += paid * float64(trade.Date.AsDays())
			}

			if amount < 0.0 {
				soldPortion := 1 - (-1.0 * (amount / totalBalance))
				fmt.Printf("Portion left over after selling: %0.3f\n", soldPortion)
				totalPaid *= soldPortion
				totalPaidByDate *= soldPortion
			}

			// Adjust balance by amount
			totalBalance += amount
		}
	}

	return trades, nil
}

//var reportTypes = []string{
//	"Text",
//	"JSON",
//	"CSV",
//}

func ShowText(b *app.BookPrinter, report []*book.Trade, ccy string) error {

	// Number of columns
	cols := 12

	// Date, Payee, Account, Amount, Balance

	// ColumnFormats
	fmts := make([]bool, 0, cols)
	fmts = append(fmts, false)
	fmts = append(fmts, true)
	fmts = append(fmts, true)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)
	fmts = append(fmts, false)

	// Rows
	rows := make([][]app.ColumnValue, 0, len(report)+1)

	// Header
	header := make([]app.ColumnValue, 0, cols)
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Date")))
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Account")))
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Payee")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Price")))
	header = append(header, app.ColumnString(b.Ansi(app.UL, "Source")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Amount")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Balance")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Value")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Cost")))
	header = append(header, app.ColumnString(b.Ansi(app.UL, "CostDate")))
	header = append(header, app.ColumnRightString(b.Ansi(app.UL, "Gain")))
	header = append(header, app.ColumnString(b.Ansi(app.UL, "GainDate")))
	rows = append(rows, header)

	// Data
	for _, t := range report {
		row := make([]app.ColumnValue, 0, cols)
		row = append(row, app.ColumnString(t.Date.String()))
		row = append(row, app.ColumnString(t.Account))
		row = append(row, app.ColumnString(t.Payee))
		row = append(row, b.GetColumnMoney(ccy, t.Price))
		row = append(row, app.ColumnString(t.PriceSource))
		row = append(row, b.GetColumnMoney(t.CCY, t.Amount))
		row = append(row, b.GetColumnMoney(t.CCY, t.Balance))
		row = append(row, b.GetColumnMoney(ccy, t.Value))
		row = append(row, b.GetColumnMoney(ccy, t.Cost))
		row = append(row, app.ColumnRightString(t.CostAvgDate.String()))
		row = append(row, b.GetColumnMoney(ccy, t.Gains))
		row = append(row, app.ColumnRightString(t.GainsAvgDate.String()))
		rows = append(rows, row)
	}

	b.PrintColumns(rows, fmts)

	return nil
}
