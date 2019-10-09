package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	// Routes
	PlaygroundsURL = "/playgrounds"
	// Other
	JsonContentType    = "application/json"
	GzipAcceptEncoding = "gzip"
)

type playgroundServer struct {
	database PlaygroundStore
	http.Handler
}

type PlaygroundStore interface {
	AllPlaygrounds() []Playground
	Playground(ID int) (Playground, error)
}

type Playground struct {
	Name      string
	Long, Lat float64
}

func New(store PlaygroundStore) *playgroundServer {
	svr := new(playgroundServer)
	svr.database = store
	router := newRouter(svr)
	svr.Handler = router
	return svr
}

func newRouter(svr *playgroundServer) *mux.Router {
	router := mux.NewRouter()

	router.Handle(PlaygroundsURL, http.HandlerFunc(svr.getAllPlaygrounds)).Methods(http.MethodGet)
	router.Handle(PlaygroundsURL+"/", http.HandlerFunc(svr.getAllPlaygrounds)).Methods(http.MethodGet)
	router.Handle(PlaygroundsURL+"/{ID}", http.HandlerFunc(svr.getPlayground)).Methods(http.MethodGet)
	// router.Handle(PlaygroundsURL+"/{ID}", http.HandlerFunc(svr.getPlayground)).Methods(http.MethodPost)

	return router
}

func (p *playgroundServer) getAllPlaygrounds(w http.ResponseWriter, r *http.Request) {
	err := encodeToJson(w, p.database.AllPlaygrounds())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *playgroundServer) getPlayground(w http.ResponseWriter, r *http.Request) {
	ID, err := extractIDFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	playground, err := p.database.Playground(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = encodeToJson(w, playground)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func extractIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	ID, err := strconv.Atoi(vars["ID"])
	if err != nil {
		return 0, fmt.Errorf("Couldn't get id from request, %s", r.URL.String())
	}
	return ID, nil
}

func encodeToJson(w http.ResponseWriter, data interface{}) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", JsonContentType)
	w.Header().Set("Accept-Encoding", GzipAcceptEncoding)
	return json.NewEncoder(w).Encode(data)
}
