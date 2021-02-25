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

func (b functionBuilder) deltaFirstDerivative(eps float64) (float64, float64) {
	f0 := (4 * eps) / math.Pow(eps+1, 2)
	return f0, -f0
}

func (b functionBuilder) deltaSecondDerivative(eps float64) (float64, float64) {
	der := -(8 * (eps - 3) * eps) / math.Pow(eps+1, 3)
	return der, der
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

func (b functionBuilder) smoothFirstDer(eps float64) (float64, float64) {
	f0 := -2 / (3 * eps * eps * (1 + math.Cosh(1/eps)))
	return f0, -f0
}

func (b functionBuilder) smoothSecondDer(eps float64) (float64, float64) {
	f0 := (-2 * (math.Tanh(1/(2*eps)) - 2*eps)) / (3 * math.Pow(eps, 4) * (math.Cosh(1/eps) + 1))
	f1 := (math.Tanh(1/(2*eps)) - 2*eps) * math.Pow(1/math.Cosh(1/(2*eps)), 2) / (3 * math.Pow(eps, 4))
	return f0, f1
}

func (b functionBuilder) customFunction(eps float64) pureFunction {
	return func(x float64) float64 {
		ch := math.Exp(x - 1)
		return (1-math.Pow(eps, 0.5)/ch)/(1+math.Exp(-16*(x-0.5))) + math.Pow(eps, 0.5)/ch*(math.Pow(x-0.5, 3)/(eps*eps*x+0.5))
	}
}

func (b functionBuilder) customFirstDer(eps float64) (float64, float64) {
	f0 := -0.340241 / math.Sqrt(eps)
	f1 := -(4 * (0.0937081 + 0.687333*math.Pow(eps, 2) + 0.499833*math.Pow(eps, 4))) / (math.Sqrt(eps) * math.Pow(1+2*math.Pow(eps, 2), 2))
	return f0, f1
}

func (b functionBuilder) customSecondDer(eps float64) (float64, float64) {
	f0 := -0.170121 / math.Pow(eps, 1.5)
	return f0, 0
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
	var step float64 = 1.0 / float64(n-1)
	for i := 0; i < n; i++ {
		fH = append(fH, f(float64(i)*step))
		xs = append(xs, float64(i)*step)
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
		return b.deltaFirstDerivative(b.eps)
	case "smooth":
		return b.smoothFirstDer(b.eps)
	case "custom":
		return b.customFirstDer(b.eps)
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
		return b.deltaSecondDerivative(b.eps)
	case "smooth":
		return b.smoothSecondDer(b.eps)
	case "custom":
		return b.customSecondDer(b.eps)
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
	case "custom":
		return b.customFunction(b.eps)
	default:
		log.Println("Unknown function requested. Returning the default function")
		return b.firstFunction(b.eps)
	}
}
