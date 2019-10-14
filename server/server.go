package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/yousseffarkhani/playground/backend2/store"

	"github.com/gorilla/mux"
)

const (
	// Routes
	PlaygroundsURL        = "/playgrounds"
	NearestPlaygroundsURL = "/nearestPlaygrounds"
	// Other
	JsonContentType    = "application/json"
	GzipAcceptEncoding = "gzip"
)

type playgroundServer struct {
	database  PlaygroundStore
	apiClient store.GeolocationClient
	http.Handler
}

type PlaygroundStore interface {
	AllPlaygrounds() store.Playgrounds
	Playground(ID int) (store.Playground, error)
}

func New(store PlaygroundStore, client store.GeolocationClient) *playgroundServer {
	svr := new(playgroundServer)
	svr.database = store
	router := newRouter(svr)
	svr.Handler = router
	svr.apiClient = client
	return svr
}

func newRouter(svr *playgroundServer) *mux.Router {
	router := mux.NewRouter()

	router.Handle(PlaygroundsURL, http.HandlerFunc(svr.getAllPlaygrounds)).Methods(http.MethodGet)
	router.Handle(PlaygroundsURL+"/", http.HandlerFunc(svr.getAllPlaygrounds)).Methods(http.MethodGet)
	router.Handle(PlaygroundsURL+"/{ID}", http.HandlerFunc(svr.getPlayground)).Methods(http.MethodGet)
	router.Handle(NearestPlaygroundsURL, http.HandlerFunc(svr.getNearestPlaygrounds)).Methods(http.MethodGet)

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

func (p *playgroundServer) getNearestPlaygrounds(w http.ResponseWriter, r *http.Request) {
	adress, err := extractAdressFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nearestPlaygrounds, err := p.database.AllPlaygrounds().FindNearestPlaygrounds(p.apiClient, adress)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = encodeToJson(w, nearestPlaygrounds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func extractAdressFromRequest(r *http.Request) (string, error) {
	queryStrings := r.URL.Query()
	adress, ok := queryStrings["adress"]
	if !ok {
		return "", fmt.Errorf("No adress paramater in request, %s", r.URL.String())
	}
	if adress[0] == "" {
		return "", fmt.Errorf("Adress paramater is empty, %s", r.URL.String())
	}
	return adress[0], nil
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
