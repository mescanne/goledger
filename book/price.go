package book

import (
	"math/big"
	"sort"
)

type prices struct {
	date Date
	val  *big.Rat
}

type priceList []prices

func (p priceList) getPrice(date Date) *big.Rat {
	minIdx := 0
	maxIdx := len(p)
	midIdx := 0

	for maxIdx > (minIdx + 1) {
		midIdx = (maxIdx-minIdx)/2 + minIdx
		if date < p[midIdx].date {
			maxIdx = midIdx
		} else if date > p[midIdx].date {
			minIdx = midIdx
		} else {
			return p[midIdx].val
		}
	}

	if date < p[minIdx].date {
		return p[minIdx].val
	}

	v := p[minIdx].val

	if minIdx < midIdx {

		// Calculate number of days!
		dateFwd := big.NewRat(int64(date.DaysSince(p[minIdx].date)), int64(p[midIdx].date.DaysSince(p[minIdx].date)))

		// New value, calculate gradient
		v = big.NewRat(0, 1)
		v.Sub(p[midIdx].val, p[minIdx].val)
		v.Mul(v, dateFwd)
		v.Add(v, p[minIdx].val)
		return v
	}

	return p[minIdx].val
}

type ccyMap struct {
	unit string
	ccy  string
}

// PriceBookBuilder is used for building a price book
type priceBookBuilder struct {
	data map[ccyMap]priceList
}

// Create a new PriceBookBuilder
func newPriceBookBuilder() *priceBookBuilder {
	return &priceBookBuilder{
		data: make(map[ccyMap]priceList),
	}
}

// Add a price for a particular date. Unit is the base currency and val is the rate
// to convert into ccy. That is <amount in unit> * val = <amount in ccy>.
func (p *priceBookBuilder) addPrice(date Date, unit string, ccy string, val *big.Rat) {
	cmap := ccyMap{
		unit: unit,
		ccy:  ccy,
	}
	v, ok := p.data[cmap]
	if !ok {
		v = make(priceList, 1, 1)
		v[0] = prices{date: date, val: val}
		p.data[cmap] = v
	} else {
		p.data[cmap] = append(v, prices{date: date, val: val})
	}
}

// Convert into a PriceBook that can be used for conversions
func (p *priceBookBuilder) build() *priceBook {
	pb := &priceBook{
		data: make(map[ccyMap]priceList),
	}
	for cmap, plist := range p.data {
		pl := make(priceList, len(plist), len(plist))
		copy(pl, plist)
		sort.SliceStable(pl, func(i, j int) bool {
			if pl[i].date < pl[j].date {
				return true
			}
			return false
		})
		pb.data[cmap] = pl
	}
	return pb
}

// PriceBook is an efficient structure for price conversions
type priceBook struct {
	data map[ccyMap]priceList
}

// Get a price for a particular date, unit, and ccy and return the rate.
//
// If unit == ccy it returns 1.
//
// If there is no data available, it returns 1.
// If the date is stale or newer than earliest date, it returns the most recent date.
// If the date is between two data points, the value is extrapolated linearly.
// TODO: Allow to negate the rates.
func (p *priceBook) getPrice(date Date, unit string, ccy string) *big.Rat {
	if unit == ccy {
		return big.NewRat(1, 1)
	}
	cmap := ccyMap{
		unit: unit,
		ccy:  ccy,
	}
	v, ok := p.data[cmap]
	if ok {
		return v.getPrice(date)
	}
	return big.NewRat(1, 1)
}
