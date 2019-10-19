package main

import (
	"net/http/httptest"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/geolocationClient"

	"github.com/yousseffarkhani/playground/backend2/server"
	"github.com/yousseffarkhani/playground/backend2/store"
	"github.com/yousseffarkhani/playground/backend2/test"
)

func TestPostPlaygroundAndGet(t *testing.T) {
	database, err := store.NewFromFile("playgrounds_test.json")
	if err != nil {
		t.Fatalf("Problem opening file, %v", err)
	}
	client := geolocationClient.APIGouvFR{}
	svr := server.New(database, client, nil)

	t.Run("Get all playgrounds SORTED by name", func(t *testing.T) {
		req := test.NewGetRequest(t, server.APIPlaygrounds)
		res := httptest.NewRecorder()

		svr.ServeHTTP(res, req)

		got, err := store.NewPlaygrounds(res.Body)
		if err != nil {
			t.Fatalf("Unable to parse response into slice, '%v'", err)
		}

		want := store.Playgrounds{
			store.Playground{
				Name: "ETABLISSEMENT FENELON",
				Long: 2.31718,
				Lat:  48.87867,
				ID:   1,
			},
			store.Playground{
				Name: "LYCEE CONDORCET",
				Long: 2.32743,
				Lat:  48.87484,
				ID:   2,
			},
			store.Playground{
				Name: "LYCEE VICTOR DURUY",
				Long: 2.31565,
				Lat:  48.8533,
				ID:   3,
			},
			store.Playground{
				Name: "TEP JARDINS SAINT PAUL",
				Long: 2.36016000,
				Lat:  48.85320000,
				ID:   4,
			},
		}

		test.AssertPlaygrounds(t, got, want)
	})
	t.Run("Get playground", func(t *testing.T) {})
	t.Run("Returns bad request if playground already exists", func(t *testing.T) {})
	t.Run("Get all playgrounds returns playgrounds ordered by proximity", func(t *testing.T) {
		req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?adress=42 avenue de Flandre Paris")
		res := httptest.NewRecorder()

		svr.ServeHTTP(res, req)

		got, err := store.NewPlaygrounds(res.Body)
		if err != nil {
			t.Fatalf("Unable to parse response into slice, '%v'", err)
		}

		want := store.Playgrounds{
			// Distance : 3,37km Temps : 55 min A pied : 4,3km
			store.Playground{
				Name: "TEP JARDINS SAINT PAUL",
				Long: 2.36016000,
				Lat:  48.85320000,
				ID:   4,
			},
			// Distance : 4,1km Temps : 49 min A pied : 3,8km
			store.Playground{
				Name: "LYCEE CONDORCET",
				Long: 2.32743,
				Lat:  48.87484,
				ID:   2,
			},
			// Distance : 4,19km Temps : 59 min A pied : 4,5km
			store.Playground{
				Name: "ETABLISSEMENT FENELON",
				Long: 2.31718,
				Lat:  48.87867,
				ID:   1,
			},
			// Distance : 5,7km Temps : 1h24 A pied : 6,5km
			store.Playground{
				Name: "LYCEE VICTOR DURUY",
				Long: 2.31565,
				Lat:  48.8533,
				ID:   3,
			},
		}

		test.AssertPlaygrounds(t, got, want)
	})
}
