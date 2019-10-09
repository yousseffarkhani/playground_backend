package geolocation

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (a *APIGouvFR) defaultify() {
	if a.
	return APIGouvFR{
		baseLatAndLongURL:   "https://api-adresse.data.gouv.fr/search/?q=",
		suffixLatAndLongURL: "&limit=1",
	}
}

func (a APIGouvFR) GetLongAndLat(adress string) (lat, long float64) {
	a.
	formattedAdress := strings.Join(strings.Fields(adress), "+")
	resp, _ := http.Get(fmt.Sprintf("%s%s%s", a.baseLatAndLongURL, formattedAdress, a.suffixLatAndLongURL))
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	// locationInfo := struct{}
	// json.NewDecoder(http.Get(fmt.Sprintf("%stext%s", baseLatAndLongURL, suffixLatAndLongURL))).Decode(&locationInfo)
	long = 2.333333
	lat = 48.866667
	return
}

type APIGouvFR struct {
	baseLatAndLongURL   string
	suffixLatAndLongURL string
}

type Item struct {
	Id    int    `json:"id"`
	Type  string `json:"type"`
	By    string `json:"by"`
	Url   string `json:"url"`
	Title string `json:"title"`
	Time  int    `json:"time"`
	Text  string `json:"text"`
}