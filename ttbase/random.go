package ttbase

import (
	"W365toFET/base"
	"math/rand/v2"
)

// AcceptRandom returns true if a broken constraint with the given weight
// should be accepted. It uses a random number in conjunction with a function
// of the weight to decide. It assumes weights are in the range 0 – 100.
func AcceptRandom(weight int) bool {
	if base.MAXWEIGHT != 100 {
		base.Warning.Printf("Weight range not 0 – 100, converting. There" +
			" may be a loss of precision.\n")
		weight = (weight * 100) / base.MAXWEIGHT
	}

	if weight <= 0 {
		return true
	}
	if weight >= 100 {
		return false
	}

	// The algorithm uses the complement of the weight.
	// Calculate p = ((100 - w0) + ((100 - w)^8 / 10^13)),
	// giving p a range of 0 – 1100, the highest value results from w0 = 0
	w := uint64(100 - weight)
	w2 := w * w
	w4 := w2 * w2
	p := w + ((w4 * w4) / 10000000000000)

	// Random acceptance function
	return p > rand.Uint64N(1100)
}
