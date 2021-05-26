package book

import (
	"math/big"
	"testing"
)

func CheckPrice(t *testing.T, pb *priceBook, date Date, unit string, ccy string, exp_rate *big.Rat, exp_type PriceType) {
	v, pt := pb.getPrice(date, unit, ccy)
	if v.Cmp(exp_rate) != 0 {
		t.Fatalf("Testing %d for %s/%s: expected rate %s, got (%s) %s\n", date, unit, ccy, exp_rate.FloatString(4), v, v.FloatString(4))
	}
	if pt != exp_type {
		t.Fatalf("Testing %d for %s/%s: expected type %s, got %s\n", date, unit, ccy, exp_type, pt)
	}
}

func TestPrice(t *testing.T) {

	pbb := newPriceBookBuilder()
	pbb.addPrice(20160101, "GBP", "USD", big.NewRat(1, 1))
	pbb.addPrice(20160201, "GBP", "USD", big.NewRat(2, 1))
	pbb.addPrice(20160301, "GBP", "USD", big.NewRat(3, 1))
	pbb.addPrice(20160401, "GBP", "USD", big.NewRat(4, 1))
	pbb.addPrice(20160501, "GBP", "USD", big.NewRat(5, 1))
	pb := pbb.build()

	// Standard before, exact first, inferred, after
	CheckPrice(t, pb, 20150101, "GBP", "USD", big.NewRat(1, 1), PriceTypeOutOfRange)
	CheckPrice(t, pb, 20160101, "GBP", "USD", big.NewRat(1, 1), PriceTypeExact)
	CheckPrice(t, pb, 20160115, "GBP", "USD", big.NewRat(45, 31), PriceTypeInferred)
	CheckPrice(t, pb, 20160501, "GBP", "USD", big.NewRat(5, 1), PriceTypeExact)
	CheckPrice(t, pb, 20160515, "GBP", "USD", big.NewRat(5, 1), PriceTypeOutOfRange)

	// Try some inverse
	CheckPrice(t, pb, 20160115, "USD", "GBP", big.NewRat(31, 45), PriceTypeInferred)
	CheckPrice(t, pb, 20160515, "USD", "GBP", big.NewRat(1, 5), PriceTypeOutOfRange)

	// Odd scenarios
	CheckPrice(t, pb, 20160515, "X", "Y", big.NewRat(1, 1), PriceTypeNone)
	CheckPrice(t, pb, 20160515, "A", "A", big.NewRat(1, 1), PriceTypeExact)

}

func TestPriceMore(t *testing.T) {
	pbb := newPriceBookBuilder()
	pbb.addPrice(20160501, "GBP", "USD", big.NewRat(1, 1))
	pbb.addPrice(20160701, "GBP", "USD", big.NewRat(2, 1))
	pbb.addPrice(20161201, "GBP", "USD", big.NewRat(3, 1))
	pbb.addPrice(20170101, "GBP", "USD", big.NewRat(4, 1))
	pb := pbb.build()

	// Try again
	CheckPrice(t, pb, 20160824, "GBP", "USD", big.NewRat(40, 17), PriceTypeInferred)
}
