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
	"time"

	"github.com/yousseffarkhani/playground/backend2/configuration"

	"github.com/markbates/goth/gothic"

	"github.com/yousseffarkhani/playground/backend2/authentication"
	"github.com/yousseffarkhani/playground/backend2/store"

	"github.com/gorilla/mux"
)

const (
	// Views
	URLHome                 = "/"
	URLLogin                = "/login"
	URLLogout               = "/logout"
	URLPlaygrounds          = "/playgrounds"
	URLPlayground           = URLPlaygrounds + "/{ID}"
	URLSubmitPlayground     = URLPlaygrounds + "/submit"
	URLSubmittedPlaygrounds = "/submittedPlaygrounds"
	URLSubmittedPlayground  = URLSubmittedPlaygrounds + "/{ID}"
	URLContact              = "/contact" // TODO

	// APIs
	APIPlaygrounds          = "/api/playgrounds"
	APIPlayground           = APIPlaygrounds + "/{ID}"
	APINearestPlaygrounds   = "/api/nearestPlaygrounds"
	APIComments             = APIPlayground + "/comments"
	APIComment              = APIComments + "/{commentID}"
	APISubmittedPlaygrounds = "/api/submittedPlaygrounds"
	APISubmittedPlayground  = APISubmittedPlaygrounds + "/{ID}"
	// Other
	JsonContentType    = "application/json"
	HtmlContentType    = "text/html; charset=utf-8"
	GzipAcceptEncoding = "gzip"
)

type PlaygroundServer struct {
	database  store.PlaygroundDatabase
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

func New(playgroundStore store.PlaygroundStore, client store.GeolocationClient, views map[string]View, middlewares map[string]Middleware) *PlaygroundServer {
	svr := new(PlaygroundServer)
	svr.database.MainPlaygroundStore = playgroundStore
	svr.database.SubmittedPlaygroundStore = &store.SubmittedPlaygroundStore{}
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
	router.Handle(URLSubmitPlayground, svr.middlewares["authorized"].ThenFunc(svr.submitPlaygroundHandler)).Methods(http.MethodGet)
	router.Handle(URLPlayground, svr.middlewares["refresh"].ThenFunc(svr.playgroundHandler)).Methods(http.MethodGet)
	router.Handle(URLSubmittedPlaygrounds, svr.middlewares["authorized"].ThenFunc(svr.submittedPlaygroundsHandler)).Methods(http.MethodGet)
	router.Handle(URLSubmittedPlayground, svr.middlewares["authorized"].ThenFunc(svr.submittedPlaygroundHandler)).Methods(http.MethodGet)
	router.Handle(URLLogin, svr.middlewares["isLogged"].ThenFunc(svr.loginHandler)).Methods(http.MethodGet)
	router.HandleFunc(URLLogout, logoutHandler).Methods(http.MethodGet)
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("static"))))
	// TODO : Put back when main.go is in /cmd file
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("../static"))))

	// Authentication
	router.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler).Methods(http.MethodGet)
	router.HandleFunc("/auth/callback/{provider}", callbackHandler)

	// API
	// Playground
	// GET
	router.HandleFunc(APIPlaygrounds, svr.getAllPlaygrounds).Methods(http.MethodGet)
	router.HandleFunc(APIPlaygrounds+"/", svr.getAllPlaygrounds).Methods(http.MethodGet)
	router.HandleFunc(APIPlayground, svr.getPlayground).Methods(http.MethodGet)
	router.HandleFunc(APINearestPlaygrounds, svr.getNearestPlaygrounds).Methods(http.MethodGet)
	router.HandleFunc(APISubmittedPlaygrounds, svr.getAllSubmittedPlaygrounds).Methods(http.MethodGet)
	// POST
	router.Handle(APISubmittedPlaygrounds, svr.middlewares["authorized"].ThenFunc(svr.submitPlayground)).Methods(http.MethodPost)
	router.Handle(APIPlaygrounds, svr.middlewares["authorized"].ThenFunc(svr.addPlayground)).Methods(http.MethodPost)
	router.Handle(APISubmittedPlayground, svr.middlewares["authorized"].ThenFunc(svr.deleteSubmittedPlayground)).Methods(http.MethodPost)

	// Comment
	// GET
	router.HandleFunc(APIComments, svr.getAllComments).Methods(http.MethodGet)
	router.HandleFunc(APIComment, svr.getComment).Methods(http.MethodGet)
	// POST
	router.Handle(APIComments, svr.middlewares["authorized"].ThenFunc(svr.addComment)).Methods(http.MethodPost)
	// DELETE
	router.Handle(APIComment, svr.middlewares["authorized"].ThenFunc(svr.deleteComment)).Methods(http.MethodDelete)

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

func (p *PlaygroundServer) addComment(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		username := claims.Username
		ID, err := extractIDFromRequest(r, "ID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = r.ParseForm()
		if err != nil {
			log.Println("Couldn't parse request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newComment := store.Comment{
			Content:          strings.TrimSpace(r.FormValue("comment")),
			Author:           username,
			TimeOfSubmission: time.Now(),
		}

		err = p.database.MainPlaygroundStore.AddComment(ID, newComment)

		if err != nil {
			log.Printf("Problème à l'ajout du commentaire, %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
func (p *PlaygroundServer) deleteComment(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		playgroundID, err := extractIDFromRequest(r, "ID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		commentID, err := extractIDFromRequest(r, "commentID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = p.database.MainPlaygroundStore.DeleteComment(playgroundID, commentID, claims.Username)
		if err != nil {
			log.Printf("Impossible de supprimer le commentaire, %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

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
	p.renderView(w, r, "playgrounds", p.database.MainPlaygroundStore.AllPlaygrounds())
}

func (p *PlaygroundServer) submittedPlaygroundsHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "submittedPlaygrounds", p.database.SubmittedPlaygroundStore.AllPlaygrounds())
}

func (p *PlaygroundServer) submitPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "submitPlayground", nil)
}

func (p *PlaygroundServer) submittedPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := extractIDFromRequest(r, "ID")
	if err != nil {
		log.Println("Couldn't parse request parameter")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	playground, err := p.database.SubmittedPlaygroundStore.Playground(ID)
	switch err {
	case store.ErrorNotFoundPlayground:
		p.renderView(w, r, "404", nil)
	case nil:
		p.renderView(w, r, "submittedPlayground", playground)
	default:
		p.renderView(w, r, "internal error", nil)
	}
}

func (p *PlaygroundServer) playgroundHandler(w http.ResponseWriter, r *http.Request) {
	playground, err := p.findPlaygroundFromRequestParameter(w, r)
	switch err {
	case store.ErrorNotFoundPlayground:
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
	err := encodeToJson(w, p.database.MainPlaygroundStore.AllPlaygrounds())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) getAllSubmittedPlaygrounds(w http.ResponseWriter, r *http.Request) {
	err := encodeToJson(w, p.database.SubmittedPlaygroundStore.AllPlaygrounds())
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

	nearestPlaygrounds, err := p.database.MainPlaygroundStore.AllPlaygrounds().FindNearestPlaygrounds(p.apiClient, address)
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

func (p *PlaygroundServer) deleteSubmittedPlayground(w http.ResponseWriter, r *http.Request) {
	ID, err := extractIDFromRequest(r, "ID")

	if err != nil {
		log.Println("Couldn't parse request parameter")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p.database.SubmittedPlaygroundStore.DeletePlayground(ID)

	w.WriteHeader(http.StatusAccepted)
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

	submittedPlaygroundID, err := strconv.Atoi(formValues["ID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	submittedPlayground, err := p.database.SubmittedPlaygroundStore.Playground(submittedPlaygroundID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(formValues["longitude"], 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	latitude, err := strconv.ParseFloat(formValues["latitude"], 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newPlayground := store.Playground{
		Name:             submittedPlayground.Name,
		Address:          formValues["address"],
		PostalCode:       formValues["postal_code"],
		City:             formValues["city"],
		Department:       formValues["department"],
		Long:             longitude,
		Lat:              latitude,
		Coating:          formValues["coating"],
		Type:             formValues["type"],
		Author:           submittedPlayground.Author,
		TimeOfSubmission: submittedPlayground.TimeOfSubmission,
	}

	if formValues["open"] == "" {
		newPlayground.Open = true
	}

	errorsMap := p.database.AddPlayground(newPlayground, submittedPlaygroundID)
	if len(errorsMap) > 0 {
		log.Println(errorsMap)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (p *PlaygroundServer) submitPlayground(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		username := claims.Username
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
			Name:             formValues["name"],
			Address:          formValues["address"],
			PostalCode:       formValues["postal_code"],
			City:             formValues["city"],
			Department:       formValues["department"],
			Author:           username,
			TimeOfSubmission: time.Now(),
		}

		errorsMap := p.database.SubmitPlayground(newPlayground)
		if len(errorsMap) > 0 {
			log.Println(errorsMap)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
			log.Println(err)
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
	playground, err := p.database.MainPlaygroundStore.Playground(ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return store.Playground{}, store.ErrorNotFoundPlayground
	}
	return playground, nil
}
