package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/store"
	"github.com/yousseffarkhani/playground/backend2/test"

	"github.com/yousseffarkhani/playground/backend2/server"
)

type mockPlaygroundStore struct {
	playgrounds store.Playgrounds
}

func (m *mockPlaygroundStore) AllPlaygrounds() store.Playgrounds {
	return m.playgrounds
}

func (m *mockPlaygroundStore) Playground(ID int) (store.Playground, error) {
	if ID > len(m.playgrounds) {
		return store.Playground{}, fmt.Errorf("Playground doesn't exist")
	}
	return m.playgrounds[ID-1], nil
}

type mockGeolocationClient struct {
}

func (m *mockGeolocationClient) GetLongAndLat(adress string) (long, lat float64, err error) {
	return 2.372452, 48.886835, nil
}

func TestGet(t *testing.T) {
	// Arrange
	playground1 := store.Playground{
		Name: "test1",
		Long: 2.36016000,
		Lat:  48.85320000,
	}
	playground2 := store.Playground{
		Name: "test2",
		Long: 2.31565,
		Lat:  48.8533}
	playgrounds := store.Playgrounds{
		playground1,
		playground2,
	}
	str := &mockPlaygroundStore{playgrounds: playgrounds}
	client := &mockGeolocationClient{}
	svr := server.New(str, client)

	cases := []string{
		server.PlaygroundsURL,
		server.PlaygroundsURL + "/",
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("Get request to %s", c), func(t *testing.T) {
			// Act
			req := test.NewGetRequest(t, c)
			res := httptest.NewRecorder()
			svr.ServeHTTP(res, req)

			// Assert
			t.Run("returns status OK", func(t *testing.T) {
				assertStatusCode(t, res, http.StatusOK)
			})
			t.Run("returns a JSON Content-Type ", func(t *testing.T) {
				assertHeader(t, res, "Content-Type", server.JsonContentType)
			})
			t.Run("returns gzip Accept-Encoding", func(t *testing.T) {
				assertHeader(t, res, "Accept-Encoding", server.GzipAcceptEncoding)
			})
			t.Run("returns a list of playgrounds", func(t *testing.T) {
				got, err := store.NewPlaygrounds(res.Body)
				if err != nil {
					t.Fatalf("Unable to parse response into slice, '%v'", err)
				}

				test.AssertPlaygrounds(t, got, playgrounds)
			})
		})
	}
	t.Run("Get request to /playgrounds/{ID}", func(t *testing.T) {
		// Act
		req := test.NewGetRequest(t, "/playgrounds/1")
		res := httptest.NewRecorder()
		svr.ServeHTTP(res, req)

		// Assert
		t.Run("Returns status OK", func(t *testing.T) {
			assertStatusCode(t, res, http.StatusOK)
		})
		t.Run("Returns a JSON Content-Type ", func(t *testing.T) {
			assertHeader(t, res, "Content-Type", server.JsonContentType)
		})
		t.Run("Returns gzip Accept-Encoding", func(t *testing.T) {
			assertHeader(t, res, "Accept-Encoding", server.GzipAcceptEncoding)
		})
		t.Run("Returns an individual playground as a JSON", func(t *testing.T) {
			var got store.Playground
			json.NewDecoder(res.Body).Decode(&got)
			test.AssertPlayground(t, got, playground1)
		})
		t.Run("Returns 404 if playground doesn't exist", func(t *testing.T) {
			// Act
			req := test.NewGetRequest(t, "/playgrounds/3")
			res := httptest.NewRecorder()
			svr.ServeHTTP(res, req)

			// Assert
			assertStatusCode(t, res, http.StatusNotFound)
		})
	})

	t.Run("Get request to /nearestPlaygrounds", func(t *testing.T) {
		t.Run("Returns a list of playgrounds ordered by proximity", func(t *testing.T) {
			req := test.NewGetRequest(t, server.NearestPlaygroundsURL+"?adress=42 Avenue de Flandre Paris")
			res := httptest.NewRecorder()

			svr.ServeHTTP(res, req)

			assertStatusCode(t, res, http.StatusOK)
			assertHeader(t, res, "Content-Type", server.JsonContentType)
			assertHeader(t, res, "Accept-Encoding", server.GzipAcceptEncoding)

			got, err := store.NewPlaygrounds(res.Body)
			if err != nil {
				t.Fatalf("Unable to parse response into slice, '%v'", err)
			}

			test.AssertPlaygrounds(t, got, playgrounds)
		})
		t.Run("Returns bad request if no adress parameter in query", func(t *testing.T) {
			req := test.NewGetRequest(t, server.NearestPlaygroundsURL+"?test=42 Avenue de Flandre Paris")
			res := httptest.NewRecorder()

			svr.ServeHTTP(res, req)

			assertStatusCode(t, res, http.StatusBadRequest)
		})
		t.Run("Returns bad request if empty adress parameter", func(t *testing.T) {
			req := test.NewGetRequest(t, server.NearestPlaygroundsURL+"?adress=")
			res := httptest.NewRecorder()

			svr.ServeHTTP(res, req)

			assertStatusCode(t, res, http.StatusBadRequest)
		})
	})

	t.Run("Get request to /test returns 404", func(t *testing.T) {
		// Arrange
		svr := server.New(nil, nil)

		// Act
		req := test.NewGetRequest(t, "/test")
		res := httptest.NewRecorder()
		svr.ServeHTTP(res, req)

		// Assert
		assertStatusCode(t, res, http.StatusNotFound)
	})
}

func TestPost(t *testing.T) {
	t.Run("Post request to /playgrounds returns 400", func(t *testing.T) {
		// Arrange
		svr := server.New(nil, nil)

		// Act
		req, err := http.NewRequest(http.MethodPost, "/playgrounds", nil)
		if err != nil {
			t.Fatalf("Couldn't create request, %v", err)
		}
		res := httptest.NewRecorder()
		svr.ServeHTTP(res, req)

		// Assert
		assertStatusCode(t, res, http.StatusMethodNotAllowed)
	})
}

func assertStatusCode(t *testing.T, res *httptest.ResponseRecorder, want int) {
	t.Helper()
	got := res.Code
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func assertHeader(t *testing.T, res *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	got := res.Header().Get(key)
	if got != want {
		t.Errorf("Got %q, want %q", got, want)
	}
}
