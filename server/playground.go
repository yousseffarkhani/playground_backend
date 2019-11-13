package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

var ErrEmptyField = errors.New("Empty field")

func (p *PlaygroundServer) getAllPlaygroundsHandler(w http.ResponseWriter, r *http.Request) {
	playgrounds := p.database.GetAllPlaygrounds()
	playgrounds.SortByName()
	err := encodeToJson(w, playgrounds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) getPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := extractParameterFromRequest(r, "ID")
	if err != nil {
		log.Println("Couldn't parse request parameter")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	playground := p.database.GetPlayground(ID)
	if playground.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = encodeToJson(w, playground)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) getNearestPlaygroundsHandler(w http.ResponseWriter, r *http.Request) {
	address, err := extractAddressFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	playgrounds := p.database.GetAllPlaygrounds()
	nearestPlaygrounds, err := playgrounds.FindNearestPlaygrounds(p.apiClient, address)
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

func (p *PlaygroundServer) addPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		r.ParseForm()
		var err error
		newPlayground := Playground{Draft: false}
		newPlayground, err = createPlaygroundFromForm(newPlayground, r.Form, claims.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if isAlreadyExisting(newPlayground, p.database.GetAllPlaygrounds()) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = p.database.AddPlayground(newPlayground)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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

func isAlreadyExisting(newPlayground Playground, playgrounds Playgrounds) bool {
	if isNameOrAddressAlreadyExisting(newPlayground, playgrounds) {
		return true
	}
	for _, playground := range playgrounds {
		if playground.Long == newPlayground.Long && playground.Lat == newPlayground.Lat {
			return true
		}
	}
	return false
}
