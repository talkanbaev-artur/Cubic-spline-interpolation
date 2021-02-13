package algorithm

import (
	"log"
	"math"
)

type functionBuilder struct {
	f   string
	eps float64
}

func NewFunctionBuilder(fun string, eps float64) functionBuilder {
	return functionBuilder{f: fun, eps: eps}
}

func (b functionBuilder) firstFunction(eps float64) pureFunction {
	return func(x float64) float64 {
		return (1 - math.Exp(-x/eps)) / (1 - math.Exp(-1/eps))
	}
}

func (b functionBuilder) firstFuncFirstDerivative(eps float64) (float64, float64) {
	f1 := 1 / ((math.Exp(1/eps) - 1) * eps)
	f0 := f1 + 1/eps
	return f0, f1
}

const epsInFour = 0.0625

func (b functionBuilder) firstFuncSecondDerivative(eps float64) (float64, float64) {
	ep := eps * eps
	f1 := 1 / (ep - ep*math.Exp(1/eps))

	var dif float64
	if eps >= epsInFour {
		dif = math.Exp(-1 / eps)
	}
	f0 := -1 / (ep * (1 - dif))
	return f0, f1
}

func (b functionBuilder) secondFunction(eps float64) pureFunction {
	return func(x float64) float64 {
		inv := 1 / math.Sqrt(eps)
		return 1 - ((math.Exp(-x*inv) + math.Exp((x-1)*inv)) / (1 + math.Exp(-1*inv)))
	}
}

func (b functionBuilder) secondFuncFirstDer(eps float64) (float64, float64) {
	root := 1 / math.Sqrt(eps)
	f0 := (math.Tanh(0.5 * root)) * root
	return f0, -f0
}

func (b functionBuilder) secondFuncSecondDer(eps float64) (float64, float64) {
	return -1 / eps, -1 / eps
}

func (b functionBuilder) deltaRegularistaion(eps float64) pureFunction {
	return func(x float64) float64 {
		return eps / (eps + math.Pow(2*x-1, 2))
	}
}

func (b functionBuilder) smoothStep(eps float64) pureFunction {
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

func (b functionBuilder) GetFirstDer() (float64, float64) {
	switch b.f {
	case "first":
		return b.firstFuncFirstDerivative(b.eps)
	case "second":
		return b.secondFuncFirstDer(b.eps)
	case "delta":
		return 0, 0
	case "smooth":
		return 0, 0
	default:
		return 0, 0
	}
}

func (b functionBuilder) GetSecondDer() (float64, float64) {
	switch b.f {
	case "first":
		return b.firstFuncSecondDerivative(b.eps)
	case "second":
		return b.secondFuncSecondDer(b.eps)
	case "delta":
		return 0, 0
	case "smooth":
		return 0, 0
	default:
		return 0, 0
	}
}

func (b functionBuilder) GetFunc() pureFunction {
	switch b.f {
	case "first":
		return b.firstFunction(b.eps)
	case "second":
		return b.secondFunction(b.eps)
	case "delta":
		return b.deltaRegularistaion(b.eps)
	case "smooth":
		return b.smoothStep(b.eps)
	default:
		log.Println("Unknown function requested. Returning the default function")
		return b.firstFunction(b.eps)
	}
}
