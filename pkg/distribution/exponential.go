package distribution

import "gonum.org/v1/gonum/stat/distuv"

type ExponentialGenerator struct {
	Generator *distuv.Exponential
}

func exponentialGenerator(rate float64) *ExponentialGenerator {
	return &ExponentialGenerator{
		Generator: &distuv.Exponential{Rate: rate, Src: nil},
	}
}

func (e *ExponentialGenerator) getNext() int64 {
	return int64(e.Generator.Rand())
}
