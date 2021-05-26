package book

import (
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

	// Conversions to base
	BaseCCY     string    // BaseCCY (always the same)
	BaseAmount  *big.Rat  // Amount in BaseCCY
	BaseSource  PriceType // Source of price for conversion
	BaseBalance *big.Rat  // Total balance to date across all CCYs converted to BaseCCY (recalculated each time)
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

func (b *Book) ExtractRegister(baseccy string, re *regexp.Regexp, split bool) RegistryReport {
	data := make([]*RegistryEntry, 0, 100)
	bals := make(map[string]*big.Rat)
	baseBal := big.NewRat(0, 1)

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

				// Calculate baseAmt, baseSource, and baseBal
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
				baseBal = baseBal.Add(baseBal, baseAmt)

				// Update data
				data = append(data, &RegistryEntry{
					Date:           trans.GetDate(),
					Payee:          trans.GetPayee(),
					Account:        p.GetAccount(),
					CounterAccount: v,
					Amount:         camts[i],
					CCY:            p.GetCCY(),
					Balance:        big.NewRat(0, 1).Set(bal),
					BaseCCY:        baseccy,
					BaseAmount:     baseAmt,
					BaseSource:     baseSource,
				})
			}
		}
	}

	return data
}
