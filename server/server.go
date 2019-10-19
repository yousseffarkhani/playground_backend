package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/yousseffarkhani/playground/backend2/store"

	"github.com/gorilla/mux"
)

const (
	// Views
	URLHome          = "/"
	URLLogin         = "/login"
	URLLogout        = "/logout"
	URLPlaygrounds   = "/playgrounds"
	URLPlayground    = URLPlaygrounds + "/{ID}"
	URLNearest       = "/nearest"
	URLContact       = "/contact"
	URLAddPlayground = "/playgrounds"

	// Routes
	APIPlaygrounds        = "/api/playgrounds"
	APINearestPlaygrounds = "/api/nearestPlaygrounds"
	// Other
	JsonContentType    = "application/json"
	HtmlContentType    = "text/html; charset=utf-8"
	GzipAcceptEncoding = "gzip"
)

type playgroundServer struct {
	database  PlaygroundStore
	apiClient store.GeolocationClient
	http.Handler
	views map[string]View
}

type View interface {
	Render(w io.Writer, r *http.Request, data interface{}) error
}

type PlaygroundStore interface {
	AllPlaygrounds() store.Playgrounds
	Playground(ID int) (store.Playground, error)
}

func New(store PlaygroundStore, client store.GeolocationClient, views map[string]View) *playgroundServer {
	svr := new(playgroundServer)
	svr.database = store
	router := newRouter(svr)
	svr.Handler = router
	svr.apiClient = client
	svr.views = views
	return svr
}

func newRouter(svr *playgroundServer) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/sw.js", serveSW).Methods(http.MethodGet)

	// Views
	router.Handle(URLHome, http.HandlerFunc(svr.homeHandler)).Methods(http.MethodGet)
	router.Handle(URLPlaygrounds, http.HandlerFunc(svr.playgroundsHandler)).Methods(http.MethodGet)
	router.Handle(URLPlayground, http.HandlerFunc(svr.playgroundHandler)).Methods(http.MethodGet)
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	// TODO : Put back when main.go is in /cmd file
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("../static"))))

	// API
	router.Handle(APIPlaygrounds, http.HandlerFunc(svr.getAllPlaygrounds)).Methods(http.MethodGet)
	router.Handle(APIPlaygrounds+"/", http.HandlerFunc(svr.getAllPlaygrounds)).Methods(http.MethodGet)
	router.Handle(APIPlaygrounds+"/{ID}", http.HandlerFunc(svr.getPlayground)).Methods(http.MethodGet)
	router.Handle(APINearestPlaygrounds, http.HandlerFunc(svr.getNearestPlaygrounds)).Methods(http.MethodGet)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, URLHome, http.StatusFound)
	}).Methods(http.MethodGet)

	return router
}

func serveSW(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "sw.js")
}

func (p *playgroundServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "home", nil)
}

func (p *playgroundServer) playgroundsHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "playgrounds", p.database.AllPlaygrounds())
}

func (p *playgroundServer) playgroundHandler(w http.ResponseWriter, r *http.Request) {
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
	p.renderView(w, r, "playground", playground)
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
	if len(nearestPlaygrounds) > 10 {
		nearestPlaygrounds = nearestPlaygrounds[:10]
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
	w.Header().Set("Content-Type", JsonContentType)
	w.Header().Set("Accept-Encoding", GzipAcceptEncoding)
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(data)
}

func (p *playgroundServer) renderView(w http.ResponseWriter, r *http.Request, template string, data interface{}) {
	if view, ok := p.views[template]; ok {
		w.Header().Set("Content-Type", HtmlContentType)
		w.Header().Set("Accept-Encoding", GzipAcceptEncoding)
		w.WriteHeader(http.StatusOK)
		err := view.Render(w, r, data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}
