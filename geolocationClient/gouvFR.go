package geolocationClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (a *APIGouvFR) defaultify() {
	if a.ApiBase == "" {
		a.ApiBase = "https://api-adresse.data.gouv.fr/search/?q="
	}
	if a.ApiSuffix == "" {
		a.ApiSuffix = "&limit=1"
	}
}

func (a APIGouvFR) GetLongAndLat(adress string) (float64, float64, error) {
	a.defaultify()

	formattedAdress := strings.Join(strings.Fields(adress), "+")

	var info GeolocationInfo

	resp, err := http.Get(fmt.Sprintf("%s%s%s", a.ApiBase, formattedAdress, a.ApiSuffix))
	if err != nil {
		return 0, 0, fmt.Errorf("Couldn't get info, %s", err)
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return 0, 0, fmt.Errorf("Couldn't parse response, %s", err)
	}
	if len(info.Features) == 0 {
		return 0, 0, fmt.Errorf("Empty answer from the API, %s", err)
	}

	long := info.Features[0].Geometry.Coordinates[0]
	lat := info.Features[0].Geometry.Coordinates[1]
	fmt.Printf("Adress : %s, Long : %.6f / Lat : %.6f", adress, long, lat)

	return long, lat, nil
}

type APIGouvFR struct {
	ApiBase   string
	ApiSuffix string
}

type GeolocationInfo struct {
	Features []struct {
		Type     string `json:"type"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}
