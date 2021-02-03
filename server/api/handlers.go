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
	EvalX        []float64        `json:"S_x"`
	EvalY        []friendlyDouble `json:"S_y"`
}

func splineHandler(w http.ResponseWriter, r *http.Request) {
	var inp splineDataInput
	err := json.NewDecoder(r.Body).Decode(&inp)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to parse json request", http.StatusBadRequest)
		return
	}
	function := parseFunctionArgument(inp.Function, inp.Epsilion)
	var out splineDataOutput
	var tempDouble []float64
	out.InitialFuncX, tempDouble = algorithm.EvaluatePureFunction(function, inp.DataPointsNum)
	out.InitialFuncY = friendlyfyFloat64(tempDouble)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func parseFunctionArgument(f string, eps float64) func(x float64) float64 {
	b := algorithm.NewFunctionBuilder()
	switch f {
	case "first":
		return b.FirstFunction(eps)
	case "second":
		return b.SecondFunction(eps)
	case "delta":
		return b.DeltaRegularistaion(eps)
	case "smooth":
		return b.SmoothStep(eps)
	default:
		log.Println("Unknown function requested. Returning the default function")
		return b.FirstFunction(eps)
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
