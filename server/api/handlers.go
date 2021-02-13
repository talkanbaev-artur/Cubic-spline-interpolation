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
	var tempDouble []float64

	//get specified function
	builder := algorithm.NewFunctionBuilder(inp.Function, inp.Epsilion)
	//calculate values
	out.InitialFuncX, tempDouble = algorithm.EvaluatePureFunction(builder.GetFunc(), 10000)
	out.InitialFuncY = friendlyfyFloat64(tempDouble)

	//build second der spline
	secF0, secF1 := builder.GetSecondDer()
	splineSecondDer, _ := algorithm.NewSplineWithPrecacl(builder.GetFunc(), inp.DataPointsNum, algorithm.SecondDerivative, secF0, secF1)
	tempDouble2 := splineSecondDer.OnRange(out.InitialFuncX)
	out.SecondDer = splineData{EvalX: out.InitialFuncX, EvalY: friendlyfyFloat64(tempDouble2)}

	//build first der spline
	firF0, firF1 := builder.GetFirstDer()
	splineFirst, _ := algorithm.NewSplineWithPrecacl(builder.GetFunc(), inp.DataPointsNum, algorithm.FirstDerivative, firF0, firF1)
	tempDouble2 = splineFirst.OnRange(out.InitialFuncX)
	out.FirstDer = splineData{EvalX: out.InitialFuncX, EvalY: friendlyfyFloat64(tempDouble2)}

	//calc err
	var errF = make([]float64, len(tempDouble))
	var max float64
	for i := 0; i < len(tempDouble); i++ {
		errF[i] = math.Abs(tempDouble[i] - tempDouble2[i])
		if errF[i] > max {
			max = errF[i]
		}
	}
	out.SecondDer.Err = max
	log.Printf("Max err:%f", max)
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
