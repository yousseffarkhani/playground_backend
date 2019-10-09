package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/test"

	"github.com/yousseffarkhani/playground/backend2/server"
)

type mockPlaygroundStore struct {
	playgrounds []server.Playground
}

func (m *mockPlaygroundStore) AllPlaygrounds() []server.Playground {
	return m.playgrounds
}

func (m *mockPlaygroundStore) Playground(ID int) (server.Playground, error) {
	if ID > len(m.playgrounds) {
		return server.Playground{}, fmt.Errorf("Playground doesn't exist")
	}
	return m.playgrounds[ID-1], nil
}

func TestGet(t *testing.T) {
	// Arrange
	playground1 := server.Playground{Name: "test1"}
	playground2 := server.Playground{Name: "test2"}
	playgrounds := []server.Playground{
		playground1,
		playground2,
	}
	store := &mockPlaygroundStore{playgrounds: playgrounds}
	svr := server.New(store)

	cases := []string{
		server.PlaygroundsURL,
		server.PlaygroundsURL + "/",
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("Get request to %s", c), func(t *testing.T) {
			// Act
			req := newGetRequest(t, c)
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
				var got []server.Playground
				json.NewDecoder(res.Body).Decode(&got)
				test.AssertPlaygrounds(t, got, playgrounds)
			})
		})
	}
	t.Run("Get request to /playgrounds/{ID}", func(t *testing.T) {
		// Act
		req := newGetRequest(t, "/playgrounds/1")
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
		t.Run("Returns an individual playground as a JSON", func(t *testing.T) {
			var got server.Playground
			json.NewDecoder(res.Body).Decode(&got)
			test.AssertPlayground(t, got, playground1)
		})
		t.Run("Returns 404 if playground doesn't exist", func(t *testing.T) {
			// Act
			req := newGetRequest(t, "/playgrounds/3")
			res := httptest.NewRecorder()
			svr.ServeHTTP(res, req)

			// Assert
			assertStatusCode(t, res, http.StatusNotFound)
		})
	})

	t.Run("Get request to /test returns 404", func(t *testing.T) {
		// Arrange
		svr := server.New(nil)

		// Act
		req := newGetRequest(t, "/test")
		res := httptest.NewRecorder()
		svr.ServeHTTP(res, req)

		// Assert
		assertStatusCode(t, res, http.StatusNotFound)
	})
}

func TestPost(t *testing.T) {
	t.Run("Post request to /playgrounds returns 400", func(t *testing.T) {
		// Arrange
		svr := server.New(nil)

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

func newGetRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Couldn't create request, %v", err)
	}
	return req
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
