package book

import (
	"math/big"
	"testing"
)

func TestPrice(t *testing.T) {

	pbb := newPriceBookBuilder()

	pbb.addPrice(20160101, "GBP", "USD", big.NewRat(1, 1))
	pbb.addPrice(20160201, "GBP", "USD", big.NewRat(2, 1))
	pbb.addPrice(20160301, "GBP", "USD", big.NewRat(3, 1))
	pbb.addPrice(20160401, "GBP", "USD", big.NewRat(4, 1))
	pbb.addPrice(20160501, "GBP", "USD", big.NewRat(5, 1))
	pb := pbb.build()

	if v := pb.getPrice(20160101, "GBP", "USD"); v.Cmp(big.NewRat(1, 1)) != 0 {
		t.Fatalf("Expected 1000000, got %v\n", v)
	}
	if v := pb.getPrice(20160115, "GBP", "USD"); v.Cmp(big.NewRat(45, 31)) != 0 {
		t.Fatalf("Expected 1500000, got %v\n", v)
	}
	if v := pb.getPrice(20160515, "GBP", "USD"); v.Cmp(big.NewRat(5, 1)) != 0 {
		t.Fatalf("Expected 5000000, got %v\n", v)
	}
}
