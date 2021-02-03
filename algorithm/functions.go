package algorithm

import (
	"math"
)

type functionBuilder struct{}

func NewFunctionBuilder() functionBuilder {
	return functionBuilder{}
}

func (b functionBuilder) FirstFunction(eps float64) pureFunction {
	return func(x float64) float64 {
		return (1 - math.Exp(-x/eps)) / (1 - math.Exp(-1/eps))
	}
}
func (b functionBuilder) SecondFunction(eps float64) pureFunction {
	return func(x float64) float64 {
		inv := 1 / math.Sqrt(eps)
		return 1 - ((math.Exp(-x*inv) + math.Exp(x-1*inv)) / (1 + math.Exp(-1*inv)))
	}
}

func (b functionBuilder) DeltaRegularistaion(eps float64) pureFunction {
	return func(x float64) float64 {
		return eps / (eps + math.Pow(2*x-1, 2))
	}
}

func (b functionBuilder) SmoothStep(eps float64) pureFunction {
	return func(x float64) float64 {
		if x <= 0.5 {
			val := math.Exp((2*x - 1) / eps)
			return -2 * (val - 1) / (3 * (val + 1))
		}
		val := math.Exp((1 - 2*x) / eps)
		return -2 * (1 - val) / (3 * (1 + val))
	}
}

func generateSplineDataPoints(numberOfPoints int) []float64 {
	h := 1.0 / float64(numberOfPoints-1)
	pointSet := make([]float64, numberOfPoints)
	for i := 0; i < numberOfPoints; i++ {
		pointSet[i] = float64(i-1) * h
	}
	return pointSet
}

type pureFunction func(x float64) float64

func EvaluatePureFunction(f pureFunction, n int) (xs []float64, fH []float64) {
	var step float64 = 1.0 / float64(n)
	for i := 0.0; i <= 1.0; i += step {
		fH = append(fH, f(i))
		xs = append(xs, i)
	}
	return
}
