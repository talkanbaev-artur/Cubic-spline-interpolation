package api

import (
	"EHDW/Cubic-spline-interpolation/algorithm"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

//RegisterRoutes registers all api routes
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/spline", splineHandler)
}

//AttachSPA attaches the spa handler, which serves the js app
func AttachSPA(r *mux.Router, base string, index string) {
	h := spaHandler{base: base, index: index}
	r.PathPrefix("/").Handler(h)
}

type spaHandler struct {
	base  string
	index string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	path = filepath.Join(h.base, path)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.base, h.index))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.base)).ServeHTTP(w, r)
}

type splineDataInput struct {
	DataPointsNum int     `json:"points"`
	Epsilion      float64 `json:"eps"`
	Function      string  `json:"func"`
}

type friendlyDouble float64

type splineDataOutput struct {
	InitialFuncX []float64        `json:"f_x"`
	InitialFuncY []friendlyDouble `json:"f_h"`
	FirstDer     splineData       `json:"first_spline"`
	SecondDer    splineData       `json:"second_spline"`
	NormalSpline splineData       `json:"normal_spline"`
}

type splineData struct {
	EvalX []float64        `json:"x"`
	EvalY []friendlyDouble `json:"y"`
	Err   float64          `json:"err"`
}

func splineHandler(w http.ResponseWriter, r *http.Request) {
	//decode
	var inp splineDataInput
	err := json.NewDecoder(r.Body).Decode(&inp)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to parse json request", http.StatusBadRequest)
		return
	}
	var out splineDataOutput
	var originalVals []float64

	//get specified function
	builder := algorithm.NewFunctionBuilder(inp.Function, inp.Epsilion)
	//calculate values
	out.InitialFuncX, originalVals = algorithm.EvaluatePureFunction(builder.GetFunc(), 10001)
	out.InitialFuncY = friendlyfyFloat64(originalVals)

	var tempDouble2 []float64
	//build second der spline
	secF0, secF1 := builder.GetSecondDer()
	splineSecondDer, _ := algorithm.NewSplineWithPrecacl(builder.GetFunc(), inp.DataPointsNum, algorithm.SecondDerivative, secF0, secF1)
	tempDouble2 = splineSecondDer.OnRange(out.InitialFuncX)
	out.SecondDer = splineData{EvalX: out.InitialFuncX, EvalY: friendlyfyFloat64(tempDouble2)}
	out.SecondDer.Err = calculateError(builder.GetFunc(), splineSecondDer, inp.DataPointsNum)
	//log.Printf("%f : %f", 0.25, splineSecondDer.Eval(0.25))

	var tempDouble3 []float64
	//build first der spline
	firF0, firF1 := builder.GetFirstDer()
	splineFirst, _ := algorithm.NewSplineWithPrecacl(builder.GetFunc(), inp.DataPointsNum, algorithm.FirstDerivative, firF0, firF1)
	tempDouble3 = splineFirst.OnRange(out.InitialFuncX)
	out.FirstDer = splineData{EvalX: out.InitialFuncX, EvalY: friendlyfyFloat64(tempDouble3)}
	out.FirstDer.Err = calculateError(builder.GetFunc(), splineFirst, inp.DataPointsNum)

	var tempDouble4 []float64
	//normal spline
	splineNormal, _ := algorithm.NewSplineWithPrecacl(builder.GetFunc(), inp.DataPointsNum, algorithm.NormalSpline, 0, 0)
	tempDouble4 = splineNormal.OnRange(out.InitialFuncX)
	out.NormalSpline = splineData{EvalX: out.InitialFuncX, EvalY: friendlyfyFloat64(tempDouble4)}
	out.NormalSpline.Err = calculateError(builder.GetFunc(), splineNormal, inp.DataPointsNum)

	// encode
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (out friendlyDouble) MarshalJSON() ([]byte, error) {
	var s string
	switch {
	case math.IsInf(float64(out), 1):
		s = "+inf"
	case math.IsInf(float64(out), -1):
		s = "-inf"
	case math.IsNaN(float64(out)):
		s = "NaN"
	default:
		return json.Marshal(float64(out))
	}
	return json.Marshal(s)
}

func friendlyfyFloat64(l []float64) []friendlyDouble {
	res := make([]friendlyDouble, len(l))
	for i, v := range l {
		res[i] = friendlyDouble(v)
	}
	return res
}

func calculateError(f func(float64) float64, s *algorithm.Spline, n int) float64 {
	ys := make([]float64, n-1)
	var h float64 = 1.0 / float64(n-1)
	for i := 0; i < n-1; i++ {
		in := float64(i)
		ys[i] = (in*h + (in+1)*h) / 2
	}
	var max float64
	for _, y := range ys {
		err := math.Abs(f(y) - s.Eval(y))
		if err > max {
			max = err
		}
	}
	return max
}
