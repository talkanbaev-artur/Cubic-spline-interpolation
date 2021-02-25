package algorithm

import (
	"errors"
	"github.com/ready-steady/linear/system"
	"log"
	_ "log"
	"math"
	"sort"
)

type boundry uint

const (
	SecondDerivative boundry = iota
	FirstDerivative
	NormalSpline
)

type Spline struct {
	gridPoints []float64 //x
	girdVals   []float64 //y
	boundry
	weights   []float64 //M
	gridSizes float64   //h
	n         int
	f0, f1    float64
}

func NewSpline(x, y []float64, b boundry, f0, f1 float64) (*Spline, error) {
	if len(x) != len(y) {
		return nil, errors.New("The data is invalid. The grid function reflection is invalid")
	}

	if !sort.Float64sAreSorted(x) {
		return nil, errors.New("The grid points are unsorted")
	}

	return &Spline{gridPoints: x, girdVals: y, boundry: b, n: len(x), f0: f0, f1: f1}, nil
}

func NewSplineWithPrecacl(f pureFunction, nPoints int, b boundry, f0, f1 float64) (*Spline, error) {
	x, y := EvaluatePureFunction(f, nPoints)
	s, err := NewSpline(x, y, b, f0, f1)
	if err != nil {
		return s, err
	}
	s.Precalc()
	return s, err
}

func (s *Spline) Precalc() {
	h := s.gridPoints[1] - s.gridPoints[0]
	s.gridSizes = h

	a := make([]float64, s.n)
	b := make([]float64, s.n)
	c := make([]float64, s.n)
	v := make([]float64, s.n)

	a[0] = 0
	c[s.n-1] = 0

	for i := 1; i < s.n-1; i++ {
		a[i] = h
		b[i] = 4 * h
		c[i] = h
		v[i] = 6 * ((s.girdVals[i+1]-s.girdVals[i])/h - (s.girdVals[i]-s.girdVals[i-1])/h)
	}

	switch s.boundry {
	case FirstDerivative:
		b[0] = 2 * h
		c[0] = h
		v[0] = 6 * ((s.girdVals[1]-s.girdVals[0])/h - s.f0)
		a[s.n-1] = h
		b[s.n-1] = 2 * h
		v[s.n-1] = 6 * (s.f1 - (s.girdVals[s.n-1]-s.girdVals[s.n-2])/h)
	case SecondDerivative:
		b[0] = 1
		c[0] = 0
		v[0] = s.f0
		a[s.n-1] = 0
		b[s.n-1] = 1
		v[s.n-1] = s.f1
	case NormalSpline:
		b[0] = 1
		c[0] = 0
		v[0] = 0
		a[s.n-1] = 0
		b[s.n-1] = 1
		v[s.n-1] = 0
	}

	slopes := system.ComputeTridiagonal(a, b, c, v)
	slopes = s.systemSolve(a, b, c, v)
	//slopes = make([]float64, len(slopes))
	//for i := range slopes {
	//	slopes[i] = 1
	//}
	s.weights = slopes
}

func (s *Spline) Eval(x float64) float64 {
	i := int(math.Floor(x / s.gridSizes))
	if i == len(s.gridPoints)-1 {
		i--
	}
	//log.Printf("x:%f, i:%d", x, i)
	//log.Printf("h[i]=%f, x[i]=%f, x[i+1]=%f", s.gridSizes[i], s.gridPoints[i], s.gridPoints[i+1])
	f1 := math.Pow(x-s.gridPoints[i], 3) / (6 * s.gridSizes) * s.weights[i+1]
	f2 := math.Pow(s.gridPoints[i+1]-x, 3) / (6 * s.gridSizes) * s.weights[i]
	f3 := (s.girdVals[i+1] - (s.gridSizes * s.gridSizes * s.weights[i+1] / 6)) * ((x - s.gridPoints[i]) / s.gridSizes)
	f4 := (s.girdVals[i] - (s.gridSizes * s.gridSizes * s.weights[i] / 6)) * ((s.gridPoints[i+1] - x) / s.gridSizes)
	return f1 + f2 + f3 + f4
}

func (s *Spline) OnRange(x []float64) []float64 {
	y := make([]float64, len(x))
	for i, xi := range x {
		y[i] = s.Eval(xi)
	}
	return y
}

func (s Spline) systemSolve(a, b, c, d []float64) []float64 {
	var n = len(b)
	if len(a) != n || len(c) != n || len(d) != n {
		log.Panic("Wrong length")
	}
	var alph, bt = make([]float64, n-1), make([]float64, n-1)
	alph[0] = -c[0] / b[0]
	bt[0] = d[0] / b[0]

	for i := 1; i < n-1; i++ {
		alph[i] = -c[i] / (b[i] - alph[i-1]*-a[i])
		bt[i] = (d[i] + bt[i-1]*-a[i]) / (b[i] - alph[i-1]*-a[i])
	}
	var u = make([]float64, n)
	u[n-1] = (d[n-1] + bt[n-2]*-a[n-1]) / (b[n-1] - alph[n-2]*-a[n-1])
	for i := n - 2; i >= 0; i-- {
		u[i] = alph[i]*u[i+1] + bt[i]
	}
	return u
}
