package algorithm

import (
	"errors"
	"sort"
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
}

func NewSpline(x, y []float64, b boundry) (*spline, error) {
	if len(x) != len(y) {
		return nil, errors.New("The data is invalid. The grid function reflection is invalid")
	}

	if !sort.Float64sAreSorted(x) {
		return nil, errors.New("The grid points are unsorted")
	}

	return &spline{gridPoints: x, girdVals: y, boundry: b, n: len(x)}, nil
}

func NewSplineWithPrecacl(x, y []float64, b boundry) (*spline, error) {
	s, err := NewSpline(x, y, b)
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
}
