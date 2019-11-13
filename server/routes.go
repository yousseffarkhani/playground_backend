package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
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
)

func newRouter(svr *PlaygroundServer) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/sw.js", serveSWHandler).Methods(http.MethodGet)

	// Views
	router.Handle(URLHome, svr.middlewares["refresh"].ThenFunc(svr.homeHandler)).Methods(http.MethodGet)
	router.Handle(URLPlaygrounds, svr.middlewares["refresh"].ThenFunc(svr.playgroundsHandler)).Methods(http.MethodGet)
	router.Handle(URLSubmitPlayground, svr.middlewares["authorized"].ThenFunc(svr.newSubmittedPlaygroundHandler)).Methods(http.MethodGet)
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
	router.HandleFunc(APIPlaygrounds, svr.getAllPlaygroundsHandler).Methods(http.MethodGet)
	router.HandleFunc(APIPlaygrounds+"/", svr.getAllPlaygroundsHandler).Methods(http.MethodGet)
	router.HandleFunc(APIPlayground, svr.getPlaygroundHandler).Methods(http.MethodGet)
	router.HandleFunc(APINearestPlaygrounds, svr.getNearestPlaygroundsHandler).Methods(http.MethodGet)
	router.HandleFunc(APISubmittedPlaygrounds, svr.getAllSubmittedPlaygroundsHandler).Methods(http.MethodGet)
	// POST
	router.Handle(APISubmittedPlaygrounds, svr.middlewares["authorized"].ThenFunc(svr.submitPlaygroundHandler)).Methods(http.MethodPost)
	router.Handle(APIPlaygrounds, svr.middlewares["authorized"].ThenFunc(svr.addPlaygroundHandler)).Methods(http.MethodPost)
	router.Handle(APISubmittedPlayground, svr.middlewares["authorized"].ThenFunc(svr.deleteSubmittedPlaygroundHandler)).Methods(http.MethodPost)

	// Comment
	// GET
	router.HandleFunc(APIComments, svr.getAllCommentsHandler).Methods(http.MethodGet)
	router.HandleFunc(APIComment, svr.getCommentHandler).Methods(http.MethodGet)
	// POST
	router.Handle(APIComments, svr.middlewares["authorized"].ThenFunc(svr.addCommentHandler)).Methods(http.MethodPost)
	// DELETE
	router.Handle(APIComment, svr.middlewares["authorized"].ThenFunc(svr.deleteCommentHandler)).Methods(http.MethodDelete)
	// PUT
	router.Handle(APIComment, svr.middlewares["authorized"].ThenFunc(svr.modifyCommentHandler)).Methods(http.MethodPut)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, URLHome, http.StatusFound)
	}).Methods(http.MethodGet)

	return router
}
