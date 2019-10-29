package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/yousseffarkhani/playground/backend2/configuration"

	"github.com/markbates/goth/gothic"

	"github.com/yousseffarkhani/playground/backend2/authentication"
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
	URLAddPlayground = URLPlaygrounds + "/add"
	URLContact       = "/contact" // TODO

	// APIs
	APIPlaygrounds        = "/api/playgrounds"
	APIPlayground         = APIPlaygrounds + "/{ID}"
	APINearestPlaygrounds = "/api/nearestPlaygrounds"
	APIComments           = APIPlayground + "/comments"
	APIComment            = APIComments + "/{commentID}"
	// Other
	JsonContentType    = "application/json"
	HtmlContentType    = "text/html; charset=utf-8"
	GzipAcceptEncoding = "gzip"
)

var ErrPlaygroundNotFound = errors.New("Playground not found")

type PlaygroundServer struct {
	database  PlaygroundStore
	apiClient store.GeolocationClient
	http.Handler
	views       map[string]View
	middlewares map[string]Middleware
}

type Middleware interface {
	ThenFunc(finalPage func(http.ResponseWriter, *http.Request)) http.Handler
}

type View interface {
	Render(w io.Writer, r *http.Request, data RenderingData) error
}

type PlaygroundStore interface {
	AllPlaygrounds() store.Playgrounds
	Playground(ID int) (store.Playground, error)
	NewPlayground(newPlayground store.Playground) map[string]error
	// AddComment(playgroundID int, comment store.Comment)
}

func New(store PlaygroundStore, client store.GeolocationClient, views map[string]View, middlewares map[string]Middleware) *PlaygroundServer {
	svr := new(PlaygroundServer)
	svr.database = store
	svr.apiClient = client
	svr.views = views
	svr.middlewares = middlewares
	router := newRouter(svr)
	svr.Handler = router
	return svr
}

func newRouter(svr *PlaygroundServer) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/sw.js", serveSW).Methods(http.MethodGet)

	// Views
	router.Handle(URLHome, svr.middlewares["refresh"].ThenFunc(svr.homeHandler)).Methods(http.MethodGet)
	router.Handle(URLPlaygrounds, svr.middlewares["refresh"].ThenFunc(svr.playgroundsHandler)).Methods(http.MethodGet)
	router.Handle(URLAddPlayground, svr.middlewares["refresh"].ThenFunc(svr.addPlaygroundHandler)).Methods(http.MethodGet)
	router.Handle(URLPlayground, svr.middlewares["refresh"].ThenFunc(svr.playgroundHandler)).Methods(http.MethodGet)
	router.HandleFunc(URLLogin, svr.loginHandler).Methods(http.MethodGet)
	router.HandleFunc(URLLogout, logoutHandler).Methods(http.MethodGet)
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	// TODO : Put back when main.go is in /cmd file
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("../static"))))

	// Authentication
	router.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler).Methods(http.MethodGet)
	router.HandleFunc("/auth/callback/{provider}", callbackHandler)

	// API
	// Playground
	router.HandleFunc(APIPlaygrounds, svr.getAllPlaygrounds).Methods(http.MethodGet)
	router.HandleFunc(APIPlaygrounds+"/", svr.getAllPlaygrounds).Methods(http.MethodGet)
	router.HandleFunc(APIPlayground, svr.getPlayground).Methods(http.MethodGet)
	router.HandleFunc(APINearestPlaygrounds, svr.getNearestPlaygrounds).Methods(http.MethodGet)
	router.HandleFunc(APIPlaygrounds, svr.addPlayground).Methods(http.MethodPost)
	// Comment
	router.HandleFunc(APIComments, svr.getAllComments).Methods(http.MethodGet)
	router.HandleFunc(APIComment, svr.getComment).Methods(http.MethodGet)
	// router.HandleFunc(APIComments, svr.addComment).Methods(http.MethodPost)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, URLHome, http.StatusFound)
	}).Methods(http.MethodGet)

	return router
}

func (p *PlaygroundServer) getAllComments(w http.ResponseWriter, r *http.Request) {
	if playground, err := p.findPlaygroundFromRequestParameter(w, r); err == nil {
		comments := playground.Comments
		if len(comments) == 0 {
			comments = store.Comments{}
		}
		err = encodeToJson(w, comments)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (p *PlaygroundServer) getComment(w http.ResponseWriter, r *http.Request) {
	if playground, err := p.findPlaygroundFromRequestParameter(w, r); err == nil {
		commentID, err := extractIDFromRequest(r, "commentID")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		comment, err := playground.FindComment(commentID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		err = encodeToJson(w, comment)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

/* func (p *PlaygroundServer) addComment(w http.ResponseWriter, r *http.Request) {
	if playground, err := p.findPlaygroundFromRequestParameter(w, r); err == nil {
		r.ParseForm()
		comment := store.Comment{
			Content: r.FormValue("comment"),
			Author:  "test",
			ID:      1,
		}
		// p.database.AddComment(playground.ID, comment)
		// playground.AddComment(comment)
		playground.Comments = append(playground.Comments, comment)
		w.WriteHeader(http.StatusAccepted)
	}
} */

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var username string
	switch {
	case user.NickName != "":
		username = user.NickName
	case user.FirstName != "":
		username = user.FirstName
	case user.Email != "":
		username = user.Email
	default:
		username = user.UserID
	}

	authentication.SetJwtCookie(w, username)
	http.Redirect(w, r, "/", http.StatusFound)
}

func serveSW(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "sw.js")
}

func (p *PlaygroundServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "home", nil)
}

func (p *PlaygroundServer) playgroundsHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "playgrounds", p.database.AllPlaygrounds())
}

func (p *PlaygroundServer) addPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "addPlayground", p.database.AllPlaygrounds())
}

func (p *PlaygroundServer) playgroundHandler(w http.ResponseWriter, r *http.Request) {
	playground, err := p.findPlaygroundFromRequestParameter(w, r)
	switch err {
	case ErrPlaygroundNotFound:
		p.renderView(w, r, "404", nil)
	case nil:
		p.renderView(w, r, "playground", playground)
	default:
		p.renderView(w, r, "internal error", nil)
	}
}

func (p *PlaygroundServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "login", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authentication.UnsetJWTCookie(w)
	http.Redirect(w, r, URLHome, http.StatusFound)
}

func (p *PlaygroundServer) getAllPlaygrounds(w http.ResponseWriter, r *http.Request) {
	err := encodeToJson(w, p.database.AllPlaygrounds())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) getPlayground(w http.ResponseWriter, r *http.Request) {
	if playground, err := p.findPlaygroundFromRequestParameter(w, r); err == nil {
		err = encodeToJson(w, playground)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (p *PlaygroundServer) getNearestPlaygrounds(w http.ResponseWriter, r *http.Request) {
	address, err := extractAddressFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nearestPlaygrounds, err := p.database.AllPlaygrounds().FindNearestPlaygrounds(p.apiClient, address)
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

func (p *PlaygroundServer) addPlayground(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	formValues := make(map[string]string)
	for key, field := range r.Form {
		if value := strings.TrimSpace(strings.Join(field, "")); value != "" {
			formValues[key] = value
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	newPlayground := store.Playground{
		Name:       formValues["name"],
		Address:    formValues["address"],
		PostalCode: formValues["postal_code"],
		City:       formValues["city"],
		Department: formValues["department"],
	}

	errorsMap := p.database.NewPlayground(newPlayground)
	if len(errorsMap) > 0 {
		log.Println(errorsMap)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func extractAddressFromRequest(r *http.Request) (string, error) {
	queryStrings := r.URL.Query()
	address, ok := queryStrings["address"]
	if !ok {
		return "", fmt.Errorf("No address paramater in request, %s", r.URL.String())
	}
	if address[0] == "" {
		return "", fmt.Errorf("Address paramater is empty, %s", r.URL.String())
	}
	return address[0], nil
}

func extractIDFromRequest(r *http.Request, parameter string) (int, error) {
	vars := mux.Vars(r)
	ID, err := strconv.Atoi(vars[parameter])
	if err != nil {
		log.Println(err)
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

type RenderingData struct {
	Username                 string
	Data                     interface{}
	GOOGLE_MAPS_API_KEY      string
	GOOGLE_GEOCODING_API_KEY string
}

func (p *PlaygroundServer) renderView(w http.ResponseWriter, r *http.Request, template string, data interface{}) {
	var username string
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		username = claims.Username
	} else {
		username = ""
	}
	renderingData := RenderingData{
		Username:                 username,
		Data:                     data,
		GOOGLE_MAPS_API_KEY:      configuration.Variables.GOOGLE_MAPS_API_KEY,
		GOOGLE_GEOCODING_API_KEY: configuration.Variables.GOOGLE_GEOCODING_API_KEY,
	}

	if view, ok := p.views[template]; ok {
		w.Header().Set("Content-Type", HtmlContentType)
		w.Header().Set("Accept-Encoding", GzipAcceptEncoding)
		err := view.Render(w, r, renderingData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (p *PlaygroundServer) findPlaygroundFromRequestParameter(w http.ResponseWriter, r *http.Request) (store.Playground, error) {
	ID, err := extractIDFromRequest(r, "ID")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return store.Playground{}, errors.New("Couldn't parse request parameter")
	}
	playground, err := p.database.Playground(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return store.Playground{}, ErrPlaygroundNotFound
	}
	return playground, nil
}
