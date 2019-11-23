package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

func (p *PlaygroundServer) getAllSubmittedPlaygroundsHandler(w http.ResponseWriter, r *http.Request) {
	playgrounds := p.database.GetAllSubmittedPlaygrounds()
	playgrounds.SortByName()
	err := encodeToJson(w, playgrounds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) submitPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		r.ParseForm()
		var err error
		newPlayground := Playground{Draft: true}
		newPlayground, err = createPlaygroundFromForm(newPlayground, r.Form, claims.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		newPlayground.ID = p.database.GetLastPlaygroundID() + 1

		if isNameOrAddressAlreadyExisting(newPlayground, p.database.GetAllSubmittedPlaygrounds()) {
			fmt.Println("Ce terrain existe ,", newPlayground.Name, newPlayground.Address)
			for _, playground := range p.database.GetAllSubmittedPlaygrounds() {
				if playground.Address == newPlayground.Address {
					fmt.Println("Ce terrain existe ???")
				}
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if isNameOrAddressAlreadyExisting(newPlayground, p.database.GetAllPlaygrounds()) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		p.database.SubmitPlayground(newPlayground)

		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *PlaygroundServer) deleteSubmittedPlaygroundHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*authentication.Claims)
	if ok {
		ID, err := extractParameterFromRequest(r, "ID")
		if err != nil {
			log.Println("Couldn't parse request parameter")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		playground := p.database.GetSubmittedPlayground(ID)
		if playground.Author != claims.Username {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = p.database.DeleteSubmittedPlayground(ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func createPlaygroundFromForm(newPlayground Playground, form url.Values, username string) (Playground, error) {
	formValues, err := trimAndVerifyEmptyField(form)
	if err != nil {
		return Playground{}, err
	}

	if _, err := strconv.Atoi(formValues["postal_code"]); err != nil {
		return Playground{}, errors.New("Postal Code should be a number")
	}
	if len(formValues["postal_code"]) != 5 {
		return Playground{}, errors.New("Postal Code should be 5 characters long")
	}
	if !newPlayground.Draft {
		newPlayground.ID, err = strconv.Atoi(formValues["ID"])
		if err != nil {
			return Playground{}, fmt.Errorf("Couldn't parse float, %s", err)
		}

		newPlayground.Long, err = strconv.ParseFloat(formValues["longitude"], 64)
		if err != nil {
			return Playground{}, fmt.Errorf("Couldn't parse float, %s", err)
		}

		newPlayground.Lat, err = strconv.ParseFloat(formValues["latitude"], 64)
		if err != nil {
			return Playground{}, fmt.Errorf("Couldn't parse float, %s", err)
		}

		if formValues["open"] == "" {
			newPlayground.Open = true
		}

		newPlayground.Type = formValues["type"]
		newPlayground.Coating = formValues["coating"]
	}

	newPlayground.Name = formValues["name"]
	newPlayground.Address = formValues["address"]
	newPlayground.PostalCode = formValues["postal_code"]
	newPlayground.City = formValues["city"]
	newPlayground.Department = formValues["department"]
	newPlayground.TimeOfSubmission = time.Now()
	newPlayground.Author = username

	return newPlayground, nil
}

func trimAndVerifyEmptyField(form url.Values) (map[string]string, error) {
	formValues := make(map[string]string)
	for key, field := range form {
		if value := strings.TrimSpace(strings.Join(field, "")); value != "" {
			formValues[key] = value
		} else {
			return nil, ErrEmptyField
		}
	}
	return formValues, nil
}

func isNameOrAddressAlreadyExisting(newPlayground Playground, playgrounds Playgrounds) bool {
	for _, playground := range playgrounds {
		if strings.ToLower(playground.Name) == strings.ToLower(newPlayground.Name) {
			return true
		}
		if strings.ToLower(playground.Address) == strings.ToLower(newPlayground.Address) {
			return true
		}
	}
	return false
}
