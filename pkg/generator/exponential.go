package generator

import "gonum.org/v1/gonum/stat/distuv"

func exponentialGenerator(rate float64) distuv.Exponential {
	return distuv.Exponential{
		Rate: rate, // Rate parameter (lambda) of the exponential distribution
		Src:  nil,  // Optional source for random number generation
	}
}

func getNext(e distuv.Exponential) {
	return e.Rand()
}
