package book

import (
	"math"
	"testing"
)

func TestXIRR(t *testing.T) {
	CheckValidIRR(t,
		[]Date{Date(20100101), Date(20110101)},
		[]float64{-1000, 1100},
		0.10)
	CheckValidIRR(t,
		[]Date{Date(20100101), Date(20110101), Date(20120101)},
		[]float64{-1000, 500, 600},
		0.0639)

	CheckValidIRR(t,
		[]Date{Date(20100101), Date(20100701), Date(20110101)},
		[]float64{-1000, 500, 600},
		0.1323)

	CheckValidIRR(t,
		[]Date{Date(20100101), Date(20110101)},
		[]float64{1000, -1100},
		0.10)

	// Matched up with Google Spreadsheet XIRR function.
	// Slightly different - likely due to date-delta calculation differences.
	dates := []Date{Date(20100101), Date(20110201), Date(20120301), Date(20130401), Date(20140101)}
	CheckValidIRR(t, dates, []float64{100, -100, -20, -30, -40}, 0.3963)
	CheckValidIRR(t, dates, []float64{-100, -100, -100, -100, 700}, 0.2468)
	CheckValidIRR(t, dates, []float64{20, 100, 120, 140, -600}, 0.2634)
}

func CheckValidIRR(t *testing.T, date []Date, cashflow []float64, result float64) {
	r := CalculateIRR(date, cashflow)
	if math.Abs(result-r) > 0.0001 {
		t.Fatalf("Expected %0.4f, got %0.4f from date %v cashflow %v", result, r, date, cashflow)
	}
}
