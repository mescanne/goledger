package book

import (
	"encoding/json"
	"math/big"
	"regexp"
	"strings"
)

type RegistryReport []*RegistryEntry
type RegistryEntry struct {

	// Registry entry
	Date           Date     `json:"date"`
	Account        string   `json:"account"`
	Payee          string   `json:"payee,omitempty"`
	CounterAccount string   `json:"counterAccount,omitempty"` // May be a single or semi-colon-delimited list
	Amount         *big.Rat `json:"amount"`
	CCY            string   `json:"ccy"`
	Balance        *big.Rat `json:"balance"` // Balance for CCY across all extract accounts
	Note           string   `json:"note"`    // Note for posting
	TNote          string   `json:"tnote"`   // Transaction note

	// Conversions to base (for transaction)
	BaseAmount *big.Rat  // Amount in BaseCCY
	BaseSource PriceType // Source of price for conversion

	// Total balance for base
	BaseBalance *big.Rat // Total balance to date across all CCYs converted to BaseCCY (recalculated each time)
}

func (rep *RegistryReport) FilterByDate(mindate, maxdate string) {
	min := DateFromString(mindate)
	max := DateFromString(maxdate)
	if min == 0 && max == 0 {
		return
	}

	startIdx := -1
	endIdx := len(*rep)
	for i, r := range *rep {
		if startIdx == -1 && (min == 0 || r.Date >= min) {
			startIdx = i
		}
		if max != 0 && r.Date >= max {
			endIdx = i
			break
		}
	}

	if startIdx == -1 {
		*rep = (*rep)[0:0]
	} else {
		*rep = (*rep)[startIdx:endIdx]
	}
}

type RegistrySummaryReport []*RegistrySummaryEntry
type RegistrySummaryEntry struct {

	// Inflow/outflow
	Date           Date
	Account        string
	Payee          string
	CounterAccount string
	CCY            string
	Cashflow       float64 // In BaseCCY

	// Total balance across entire portfolio
	Balance float64 // In BaseCCY
}

func (rep *RegistryReport) ExtractSummary(track *regexp.Regexp) RegistrySummaryReport {
	data := make([]*RegistrySummaryEntry, 0, 100)

	var bal float64 = 0.0
	for _, p := range *rep {

		// Convert
		amt, _ := p.BaseAmount.Float64()
		bal += amt

		// If tracking, then
		data = append(data, &RegistrySummaryEntry{
			Date:           p.Date,
			Account:        p.Account,
			Payee:          p.Payee,
			CounterAccount: p.CounterAccount,
			CCY:            p.CCY,
			Cashflow:       amt,
			Balance:        bal,
		})
	}

	return data
}

func (rep *RegistrySummaryReport) IRR() float64 {
	return 0.0
}

func (b *Book) ExtractRegister(baseccy string, re *regexp.Regexp, split bool) RegistryReport {
	data := make([]*RegistryEntry, 0, 100)
	bals := make(map[string]*big.Rat)
	baseBal := big.NewRat(0, 1)

	// Compact the book first
	b.compact()

	// Iterate each transaction
	for _, trans := range b.trans {

		// Iterate each posting
		for _, p := range trans {

			if !re.MatchString(p.GetAccount()) {
				continue
			}

			// Only counteraccounts of the same currency in the opposite direction are of interest!
			caccts := make([]string, 0, 1)
			camts := make([]*big.Rat, 0, 1)
			sumamt := big.NewRat(0, 1)
			for _, tp := range trans {
				if tp.GetAccount() == p.GetAccount() {
					continue
				}
				if tp.GetAmount().Sign() == p.GetAmount().Sign() {
					continue
				}
				if tp.GetCCY() != p.GetCCY() {
					continue
				}
				caccts = append(caccts, tp.GetAccount())
				camts = append(camts, big.NewRat(0, 1).Set(tp.GetAmount()))
				sumamt.Add(sumamt, tp.GetAmount())
			}

			// If we need to re-combine, do so
			if len(caccts) > 1 && !split {
				caccts[0] = strings.Join(caccts, ";")
				caccts = caccts[0:1]
				camts[0] = big.NewRat(0, 1).Set(p.GetAmount())
				camts = camts[0:1]
			} else {

				// Invert the total
				sumamt.Inv(sumamt)

				// Adjust the amounts
				for i, _ := range camts {
					camts[i] = camts[i].Mul(camts[i], p.GetAmount())
					camts[i] = camts[i].Mul(camts[i], sumamt)
				}
			}

			// Get the balance for the currency
			bal, ok := bals[p.GetCCY()]
			if !ok {
				bal = big.NewRat(0, 1)
				bals[p.GetCCY()] = bal
			}

			// Allocate them
			for i, v := range caccts {

				// Add to balance
				bal = bal.Add(bal, camts[i])

				// Calculate baseAmt and baseSource
				baseAmt := camts[i]
				var baseSource PriceType = PriceTypeExact
				if p.GetCCY() != baseccy {
					r := trans.InferRate(baseccy, p.GetCCY())
					if r != nil {
						baseSource = PriceTypeTrade
					} else {
						r, baseSource = b.GetPrice(trans.GetDate(), baseccy, p.GetCCY())
					}
					baseAmt = big.NewRat(0, 1).Mul(r, camts[i])
				}

				// Calculate the balance across all curencies

				// Update data
				data = append(data, &RegistryEntry{
					Date:           trans.GetDate(),
					Payee:          trans.GetPayee(),
					Account:        p.GetAccount(),
					CounterAccount: v,
					Amount:         camts[i],
					CCY:            p.GetCCY(),
					Note:           p.GetPostNote(),
					TNote:          p.GetTransactionNote(),
					Balance:        big.NewRat(0, 1).Set(bal),
					BaseAmount:     baseAmt,
					BaseSource:     baseSource,
					BaseBalance:    big.NewRat(0, 1).Set(baseBal),
				})
			}
		}
	}

	return data
}

func (re RegistryEntry) MarshalJSON() ([]byte, error) {

	type JsonRegistryEntry struct {
		Date           Date    `json:"date"`
		Account        string  `json:"account"`
		Payee          string  `json:"payee,omitempty"`
		CounterAccount string  `json:"counterAccount,omitempty"` // May be a single or semi-colon-delimited list
		Amount         float64 `json:"amount"`
		CCY            string  `json:"ccy"`
		Balance        float64 `json:"balance"` // Balance for CCY across all extract accounts
		Note           string  `json:"note"`    // Note for posting
		TNote          string  `json:"tnote"`   // Transaction note

		// Conversions to base
		BaseCCY     string  // BaseCCY (always the same)
		BaseAmount  float64 // Amount in BaseCCY
		BaseSource  string  // Source of price for conversion
		BaseBalance float64 // Total balance to date across all CCYs converted to BaseCCY (recalculated each time)
	}

	amt, _ := re.Amount.Float64()
	bal, _ := re.Balance.Float64()
	baseAmt, _ := re.BaseAmount.Float64()
	baseBal, _ := re.BaseBalance.Float64()

	return json.Marshal(&JsonRegistryEntry{
		Date:           re.Date,
		Account:        re.Account,
		Payee:          re.Payee,
		CounterAccount: re.CounterAccount,
		Amount:         amt,
		CCY:            re.CCY,
		Balance:        bal,
		Note:           re.Note,
		TNote:          re.TNote,

		BaseCCY:     re.BaseCCY,
		BaseAmount:  baseAmt,
		BaseSource:  re.BaseSource.String(),
		BaseBalance: baseBal,
	})
}
