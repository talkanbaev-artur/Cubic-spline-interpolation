package algorithm

import (
	"errors"
	_ "log"
	"math"
	"sort"

	"github.com/ready-steady/linear/system"
)

type boundry uint

const (
	SecondDerivative boundry = iota
	FirstDerivative
	NormalSpline
)

type spline struct {
	gridPoints []float64 //x
	girdVals   []float64 //y
	boundry
	weights   []float64 //M
	gridSizes []float64 //h
	n         int
	f0, f1    float64
}

func NewSpline(x, y []float64, b boundry, f0, f1 float64) (*spline, error) {
	if len(x) != len(y) {
		return nil, errors.New("The data is invalid. The grid function reflection is invalid")
	}

	if !sort.Float64sAreSorted(x) {
		return nil, errors.New("The grid points are unsorted")
	}

	return &spline{gridPoints: x, girdVals: y, boundry: b, n: len(x), f0: f0, f1: f1}, nil
}

func NewSplineWithPrecacl(f pureFunction, nPoints int, b boundry, f0, f1 float64) (*spline, error) {
	x, y := EvaluatePureFunction(f, nPoints)
	s, err := NewSpline(x, y, b, f0, f1)
	if err != nil {
		return s, err
	}
	s.Precalc()
	return s, err
}

func (s *spline) Precalc() {
	h := make([]float64, s.n)
	for i := 0; i < s.n-1; i++ {
		h[i] = s.gridPoints[i+1] - s.gridPoints[i]
	}
	s.gridSizes = h

	a := make([]float64, s.n)
	b := make([]float64, s.n)
	c := make([]float64, s.n)
	v := make([]float64, s.n)

	a[0] = 0
	c[s.n-1] = 0

	for i := 1; i < s.n-1; i++ {
		a[i] = h[i-1]
		b[i] = 2 * (h[i-1] + h[i])
		c[i] = h[i]
		v[i] = 6 * ((s.girdVals[i+1]-s.girdVals[i])/h[i] - (s.girdVals[i]-s.girdVals[i-1])/h[i-1])
	}

	switch s.boundry {
	case FirstDerivative:
		b[0] = 2 * h[0]
		c[0] = h[0]
		v[0] = 6 * ((s.girdVals[1]-s.girdVals[0])/h[0] - s.f0)
		a[s.n-1] = h[s.n-2]
		b[s.n-1] = 2 * h[s.n-2]
		v[s.n-1] = 6 * (s.f1 - (s.girdVals[s.n-1]-s.girdVals[s.n-2])/h[s.n-2])
	case SecondDerivative:
		b[0] = 1
		c[0] = 0
		v[0] = s.f0
		a[s.n-1] = 0
		b[s.n-1] = 1
		v[s.n-1] = s.f1
	}

	slopes := system.ComputeTridiagonal(a, b, c, v)
	//slopes = make([]float64, len(slopes))
	//for i := range slopes {
	//	slopes[i] = 1
	//}
	s.weights = slopes
}

func (s *spline) Eval(x float64) float64 {
	var i int
	for x > s.gridPoints[i] && i != len(s.gridPoints)-1 {
		i++
	}
	if i != 0 {
		i--
	}
	//log.Printf("x:%f, i:%d", x, i)
	//log.Printf("h[i]=%f, x[i]=%f, x[i+1]=%f", s.gridSizes[i], s.gridPoints[i], s.gridPoints[i+1])
	f1 := (math.Pow(x-s.gridPoints[i], 3) / (6 * s.gridSizes[i]) * s.weights[i+1])
	f2 := (math.Pow(s.gridPoints[i+1]-x, 3) / (6 * s.gridSizes[i]) * s.weights[i])
	f3 := (s.girdVals[i+1] - (s.gridSizes[i] * s.gridSizes[i] * s.weights[i+1] / 6)) * ((x - s.gridPoints[i]) / s.gridSizes[i])
	f4 := (s.girdVals[i] - (s.gridSizes[i] * s.gridSizes[i] * s.weights[i] / 6)) * ((s.gridPoints[i+1] - x) / s.gridSizes[i])
	return f1 + f2 + f3 + f4
}

func (s *spline) OnRange(x []float64) []float64 {
	y := make([]float64, len(x))
	for i, xi := range x {
		y[i] = s.Eval(xi)
	}
	return y
}
