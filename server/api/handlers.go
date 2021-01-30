package api

import (
	"encoding/json"
	"log"
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

func splineHandler(w http.ResponseWriter, r *http.Request) {
	var inp splineDataInput
	err := json.NewDecoder(r.Body).Decode(&inp)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to parse json request", http.StatusBadRequest)
		return
	}
	w.Write([]byte("Hello"))
}
