package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yousseffarkhani/playground/backend2/configuration"

	"github.com/markbates/goth/gothic"

	"github.com/yousseffarkhani/playground/backend2/authentication"

	"github.com/gorilla/mux"
)

const (
	JsonContentType    = "application/json"
	HtmlContentType    = "text/html; charset=utf-8"
	GzipAcceptEncoding = "gzip"
)

type PlaygroundServer struct {
	database  Database
	apiClient GeolocationClient
	http.Handler
	views       map[string]View
	middlewares map[string]Middleware
}

type GeolocationClient interface {
	GetLongAndLat(address string) (long, lat float64, err error)
}

type Database interface {
	PlaygroundStore
	SubmittedPlaygroundStore
	CommentStore
}

type PlaygroundStore interface {
	GetLastPlaygroundID() int
	GetAllPlaygrounds() Playgrounds
	GetPlayground(ID int) Playground
	AddPlayground(newPlayground Playground) error
}

type SubmittedPlaygroundStore interface {
	GetAllSubmittedPlaygrounds() Playgrounds
	GetSubmittedPlayground(ID int) Playground
	DeleteSubmittedPlayground(ID int) error
	SubmitPlayground(newPlayground Playground)
}

type CommentStore interface {
	GetLastCommentID() int
	GetAllComments(playgroundID int) Comments
	GetComment(playgroundID, commentID int) Comment
	AddComment(playgroundID int, newComment Comment) error
	DeleteComment(playgroundID, commentID int) error
	ModifyComment(updatedComment Comment) error
}

type Playground struct {
	Name             string    `json:"name"`
	Address          string    `json:"address"`
	PostalCode       string    `json:"postal_code"`
	City             string    `json:"city"`
	Department       string    `json:"department"`
	Long             float64   `json:"long"`
	Lat              float64   `json:"lat"`
	Coating          string    `json:"coating"`
	Type             string    `json:"type"`
	Open             bool      `json:"open"`
	ID               int       `json:"id"`
	Author           string    `json:"author"`
	TimeOfSubmission time.Time `json:"time_of_submission"`
	Draft            bool
}

type Playgrounds []Playground

type Comment struct {
	ID               int       `json:"id"`
	PlaygroundID     int       `json:"playground_id"`
	Content          string    `json:"content"`
	Author           string    `json:"author"`
	TimeOfSubmission time.Time `json:"time_of_submission"`
}

type Comments []Comment

type Middleware interface {
	ThenFunc(finalPage func(http.ResponseWriter, *http.Request)) http.Handler
}

type View interface {
	Render(w http.ResponseWriter, r *http.Request, data RenderingData) error
}

func New(database Database, apiClient GeolocationClient, views map[string]View, middlewares map[string]Middleware) *PlaygroundServer {
	svr := new(PlaygroundServer)
	svr.database = database
	svr.apiClient = apiClient
	svr.views = views
	svr.middlewares = middlewares
	router := newRouter(svr)
	svr.Handler = router
	return svr
}

func serveSWHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "sw.js")
}

func (p *PlaygroundServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "home", nil)
}

func (p *PlaygroundServer) playgroundsHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "playgrounds", p.database.GetAllPlaygrounds())
}

func (p *PlaygroundServer) playgroundHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := extractParameterFromRequest(r, "ID")
	if err != nil {
		log.Println("Couldn't parse request parameter")
		p.renderView(w, r, "internal error", nil)
		return
	}

	playground := p.database.GetPlayground(ID)
	if playground.Name == "" {
		p.renderView(w, r, "404", nil)
		return
	}

	comments := p.database.GetAllComments(ID)

	type playgroundWithComments struct {
		Playground
		Comments
	}

	p.renderView(w, r, "playground", playgroundWithComments{
		playground,
		comments,
	})
}

func (p *PlaygroundServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "login", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authentication.UnsetJWTCookie(w)
	http.Redirect(w, r, URLHome, http.StatusFound)
}

func (p *PlaygroundServer) submittedPlaygroundsHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "submittedPlaygrounds", p.database.GetAllSubmittedPlaygrounds())
}

func (p *PlaygroundServer) newSubmittedPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	p.renderView(w, r, "submitPlayground", nil)
}

func (p *PlaygroundServer) submittedPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := extractParameterFromRequest(r, "ID")
	if err != nil {
		log.Println("Couldn't parse request parameter")
		p.renderView(w, r, "internal error", nil)
		return
	}

	playground := p.database.GetSubmittedPlayground(ID)
	if playground.Name == "" {
		p.renderView(w, r, "404", nil)
		return
	}

	p.renderView(w, r, "submittedPlayground", playground)
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
		err := view.Render(w, r, renderingData)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}

func extractParameterFromRequest(r *http.Request, parameter string) (int, error) {
	vars := mux.Vars(r)
	param, err := strconv.Atoi(vars[parameter])
	if err != nil {
		log.Println(err)
		return 0, fmt.Errorf("Couldn't get parameter %q from request, %s", parameter, r.URL.String())
	}
	return param, nil
}

func (p Playgrounds) SortByName() {
	sort.Slice(p, func(i, j int) bool {
		return strings.ToLower(p[i].Name) < strings.ToLower(p[j].Name)
	})
}

func (p Playgrounds) sortByProximity(long, lat float64) Playgrounds {
	playgroundsSorted := make(Playgrounds, len(p))
	copy(playgroundsSorted, p)
	sort.SliceStable(playgroundsSorted, func(i, j int) bool {
		distanceFromAddressToI := playgroundsSorted[i].calculateSquaredDistanceFrom(long, lat)
		distanceFromAddressToJ := playgroundsSorted[j].calculateSquaredDistanceFrom(long, lat)
		return distanceFromAddressToI < distanceFromAddressToJ
	})
	return playgroundsSorted
}

func (p Playground) calculateSquaredDistanceFrom(long, lat float64) float64 {
	return math.Pow(long-p.Long, 2) + math.Pow(lat-p.Lat, 2)
}

func (p Playgrounds) FindNearestPlaygrounds(client GeolocationClient, address string) (Playgrounds, error) {
	long, lat, err := client.GetLongAndLat(address)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get longitude and lattitude, %s", err)
	}
	playgroundsSorted := p.sortByProximity(long, lat)
	return playgroundsSorted, nil
}
