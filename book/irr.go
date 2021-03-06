package book

import (
	"math"
)

func CalculateIRR(date []Date, cashflow []float64) float64 {

	// Calculate the years
	years := make([]float64, len(date))
	startYears := date[0].AsYears()
	years[0] = 0.0
	for i := 1; i < len(date); i++ {
		years[i] = date[i].AsYears() - startYears
	}

	guess := 0.10
	epsilon := 0.0001
	limit := 10000

	for i := 0; i < limit; i++ {

		// Calculate residual
		residual := cashflow[0]
		for i := 1; i < len(cashflow); i++ {
			residual += cashflow[i] / math.Pow(1+guess, years[i])
		}

		if math.Abs(residual) < epsilon {
			break
		}

		residualAdjustment := 0.0
		for i := 1; i < len(cashflow); i++ {
			residualAdjustment -= years[i] * cashflow[i] / math.Pow(1+guess, years[i]+1)
		}

		newRate := guess - residual/residualAdjustment

		if math.Abs(newRate-guess) < epsilon {
			break
		}

		if newRate < -1 {
			newRate = -0.999999999
		}
		guess = newRate
	}

	return guess
}
