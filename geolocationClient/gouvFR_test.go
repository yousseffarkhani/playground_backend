package geolocationClient_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/geolocationClient"
)

func setup() (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/42+avenue+de+Flandre+Paris&limit=1", stubGetSearch)
	svr := httptest.NewServer(mux)
	return svr.URL, svr.Close
}

func stubGetSearch(w http.ResponseWriter, r *http.Request) {
	searchResult := `{"features":[{"type": "Feature","geometry":{"type":"Point","coordinates":[2.0,3.0]}}]}`
	fmt.Fprintf(w, searchResult)
}

func TestGouvFR(t *testing.T) {
	t.Run("Get geolocation informations", func(t *testing.T) {
		URL, closeServ := setup()
		defer closeServ()
		client := geolocationClient.APIGouvFR{ApiBase: URL + "/"}

		wantLong, wantLat := 2.0, 3.0
		gotLong, gotLat, err := client.GetLongAndLat("42 avenue de Flandre Paris")
		if err != nil {
			t.Fatalf("Couldn't get geolocation info, %s", err)
		}

		if gotLong != wantLong || gotLat != wantLat {
			t.Errorf("Long and lat are not correct")
		}
	})
}

func assertAddress(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got : %q, want : %q", got, want)
	}
}
