package book

import (
	"math/big"
	"sort"
)

type Price struct {
	date Date
	val  *big.Rat
}

func (p Price) GetDate() Date {
	return p.date
}
func (p Price) GetPrice() *big.Rat {
	return p.val
}

type PriceType int

const (
	PriceTypeOutOfRange = iota
	PriceTypeInferred
	PriceTypeExact
	PriceTypeTrade
	PriceTypeNone
)

func (d PriceType) String() string {
	return [4]string{"OutOfRange", "Inferred", "Exact", "None"}[d]
}

type PriceList []Price

// Want a price iterator that
type PriceIterator struct {
	idx int
	pl  PriceList
}

func (p PriceList) GetIterator() *PriceIterator {
	return &PriceIterator{0, p}
}

func (pi *PriceIterator) GetPrice(date Date) (*big.Rat, PriceType) {

	// Move forward until date is the same or after
	for pi.idx < len(pi.pl) && pi.pl[pi.idx].date < date {
		pi.idx++
	}

	// Use last value
	if pi.idx == len(pi.pl) {
		return pi.pl[len(pi.pl)-1].val, PriceTypeOutOfRange
	}

	// Use exact value (matching date)
	if pi.pl[pi.idx].date == date {
		return pi.pl[pi.idx].val, PriceTypeExact
	}

	// Can't go prior - out of range
	if pi.idx == 0 {
		return pi.pl[0].val, PriceTypeOutOfRange
	}

	// Calculate number of days between min and max
	dateFwd := big.NewRat(int64(date.DaysSince(pi.pl[pi.idx-1].date)), int64(pi.pl[pi.idx].date.DaysSince(pi.pl[pi.idx-1].date)))

	// New value, calculate gradient
	v := big.NewRat(0, 1)
	v.Sub(pi.pl[pi.idx].val, pi.pl[pi.idx-1].val)
	v.Mul(v, dateFwd)
	v.Add(v, pi.pl[pi.idx-1].val)

	// Inferred price
	return v, PriceTypeInferred
}

func (p PriceList) GetPrice(date Date) (*big.Rat, PriceType) {

	// Check if it's before the data
	if date < p[0].date {
		return p[0].val, PriceTypeOutOfRange
	}

	// Check if it's after the data
	if date > p[len(p)-1].date {
		return p[len(p)-1].val, PriceTypeOutOfRange
	}

	// Search for the date precisely
	minIdx := 0
	maxIdx := len(p)
	for minIdx < maxIdx {
		midIdx := (minIdx + maxIdx) / 2
		if date > p[midIdx].date {
			if midIdx == minIdx {
				break
			}
			minIdx = midIdx
		} else if date < p[midIdx].date {
			maxIdx = midIdx
		} else {
			return p[midIdx].val, PriceTypeExact
		}
	}

	// Calculate number of days between min and max
	dateFwd := big.NewRat(int64(date.DaysSince(p[minIdx].date)), int64(p[minIdx+1].date.DaysSince(p[minIdx].date)))

	// New value, calculate gradient
	v := big.NewRat(0, 1)
	v.Sub(p[minIdx+1].val, p[minIdx].val)
	v.Mul(v, dateFwd)
	v.Add(v, p[minIdx].val)

	return v, PriceTypeInferred
}

type ccyMap struct {
	unit string
	ccy  string
}

// PriceBookBuilder is used for building a price book
type priceBookBuilder struct {
	data map[ccyMap]PriceList
}

// Create a new PriceBookBuilder
func newPriceBookBuilder() *priceBookBuilder {
	return &priceBookBuilder{
		data: make(map[ccyMap]PriceList),
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
		v = make(PriceList, 1, 1)
		v[0] = Price{date: date, val: val}
		p.data[cmap] = v
	} else {
		p.data[cmap] = append(v, Price{date: date, val: val})
	}
}

// Convert into a PriceBook that can be used for conversions
func (p *priceBookBuilder) build() *priceBook {
	pb := &priceBook{
		data: make(map[ccyMap]PriceList),
	}
	for cmap, plist := range p.data {
		pl := make(PriceList, len(plist), len(plist))
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
	data map[ccyMap]PriceList
}

// Get a price for a particular date, unit, and ccy and return the rate.
//
// If unit == ccy it returns 1.
//
// If there is no data available, it returns 1.
// If the date is stale or newer than earliest date, it returns the most recent date.
// If the date is between two data points, the value is extrapolated linearly.
func (p *priceBook) getPrice(date Date, unit string, ccy string) (*big.Rat, PriceType) {
	if unit == ccy {
		return big.NewRat(1, 1), PriceTypeExact
	}
	v, ok := p.data[ccyMap{unit, ccy}]
	if ok {
		return v.GetPrice(date)
	}
	v, ok = p.data[ccyMap{ccy, unit}]
	if ok {
		r, pt := v.GetPrice(date)
		return big.NewRat(0, 1).Inv(r), pt
	}
	return big.NewRat(1, 1), PriceTypeNone
}

func (p *priceBook) getPrices(unit string, ccy string) PriceList {
	v, ok := p.data[ccyMap{unit, ccy}]
	if ok {
		return v
	}
	return nil
}

func (p PriceType) merge(m PriceType) PriceType {
	if p == PriceTypeNone || m == PriceTypeNone {
		return PriceTypeNone
	}
	if p == PriceTypeOutOfRange || m == PriceTypeOutOfRange {
		return PriceTypeOutOfRange
	}
	if p == PriceTypeInferred || m == PriceTypeInferred {
		return PriceTypeInferred
	}
	if p == PriceTypeExact || m == PriceTypeExact {
		return PriceTypeExact
	}
	return PriceTypeTrade
}

func (p *priceBook) getValue(date Date, values map[string]float64, ccy string) (float64, PriceType) {
	var bal float64
	var pt PriceType = PriceTypeExact
	for unit, v := range values {
		rate, typ := p.getPrice(date, unit, ccy)
		f_rate, _ := rate.Float64()
		bal += v * f_rate
		pt = pt.merge(typ)
	}
	return bal, pt
}
