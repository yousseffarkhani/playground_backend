package server_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/server"

	"github.com/yousseffarkhani/playground/backend2/store"
	"github.com/yousseffarkhani/playground/backend2/test"
)

var playground1 = store.Playground{
	Name: "test1",
	Long: 2.36016000,
	Lat:  48.85320000,
}
var playground2 = store.Playground{
	Name: "test2",
	Long: 2.31565,
	Lat:  48.8533}
var playgrounds = store.Playgrounds{
	playground1,
	playground2,
}
var dummyMiddlewares = map[string]server.Middleware{
	"isLogged": &mockMiddleware{},
	"refresh":  &mockMiddleware{},
}

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

type mockGeolocationClient struct{}

func (m *mockGeolocationClient) GetLongAndLat(adress string) (long, lat float64, err error) {
	return 2.372452, 48.886835, nil
}

type mockView struct {
	data   server.RenderingData
	called bool
}

func (m *mockView) Render(w io.Writer, r *http.Request, data server.RenderingData) error {
	m.called = true
	m.data = data
	return nil
}

func TestViews(t *testing.T) {
	str := &mockPlaygroundStore{playgrounds: playgrounds}

	mockHomeView := &mockView{}
	mockPlaygroundsView := &mockView{}
	mockPlaygroundView := &mockView{}

	views := map[string]server.View{
		"home":        mockHomeView,
		"playgrounds": mockPlaygroundsView,
		"playground":  mockPlaygroundView,
	}

	svr := server.New(str, nil, views, dummyMiddlewares)

	type testStruct struct {
		mockView     *mockView
		expectedData server.RenderingData
	}

	tests := map[string]testStruct{
		server.URLHome: testStruct{
			mockView:     mockHomeView,
			expectedData: server.RenderingData{"", nil},
		},
		server.URLPlaygrounds: testStruct{
			mockView:     mockPlaygroundsView,
			expectedData: server.RenderingData{"", playgrounds},
		},
		server.URLPlaygrounds + "/1": testStruct{
			mockView:     mockPlaygroundView,
			expectedData: server.RenderingData{"", playground1},
		},
	}

	for url, tt := range tests {
		t.Run(fmt.Sprintf("Get to %s returns an HTML Page using correct template and data", url), func(t *testing.T) {
			req := test.NewGetRequest(t, url)
			res := httptest.NewRecorder()

			svr.ServeHTTP(res, req)

			assertStatusCode(t, res, http.StatusOK)
			assertHeader(t, res, "Content-Type", server.HtmlContentType)
			assertHeader(t, res, "Accept-Encoding", server.GzipAcceptEncoding)

			if tt.mockView.called != true {
				t.Error("View should be called")
			}
			want := tt.expectedData
			got := tt.mockView.data
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got : %v, want : %v", got, want)
			}
		})
	}

	t.Run("All URLs not matching a route REDIRECTS to home page", func(t *testing.T) {
		req := test.NewGetRequest(t, "/test")
		res := httptest.NewRecorder()

		svr.ServeHTTP(res, req)

		url, err := res.Result().Location()
		if err != nil {
			t.Fatalf("Problem with response, %s", err)
		}
		if url.Path != "/" {
			t.Errorf("URL should be \"/\", got : %q", url.Path)
		}
	})
}

func TestMiddlewares(t *testing.T) {
	mockIsLogged := &mockMiddleware{}
	mockRefreshJWT := mockIsLogged
	middlewares := map[string]server.Middleware{
		"isLogged": mockIsLogged,
		"refresh":  mockRefreshJWT,
	}
	str := &mockPlaygroundStore{}

	svr := server.New(str, nil, nil, middlewares)

	cases := []string{
		server.URLHome,
		server.URLPlaygrounds,
		server.URLPlaygrounds + "/1",
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("isLogged middleware is called on route %q", c), func(t *testing.T) {
			req := test.NewGetRequest(t, c)

			svr.ServeHTTP(httptest.NewRecorder(), req)

			if mockIsLogged.called != true {
				t.Errorf("IsLogged middleware hasn't been called")
			}
			mockIsLogged.called = false
		})
	}
}

type mockMiddleware struct {
	called bool
}

func (m *mockMiddleware) ThenFunc(finalPage func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.called = true
		finalPage(w, r)
	})
}

func TestGetAPIs(t *testing.T) {
	// Arrange
	str := &mockPlaygroundStore{playgrounds: playgrounds}
	client := &mockGeolocationClient{}

	svr := server.New(str, client, nil, dummyMiddlewares)

	cases := []string{
		server.APIPlaygrounds,
		server.APIPlaygrounds + "/",
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
	t.Run("Get request to /api/playgrounds/{ID}", func(t *testing.T) {
		// Act
		req := test.NewGetRequest(t, server.APIPlaygrounds+"/1")
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
			req := test.NewGetRequest(t, server.APIPlaygrounds+"/3")
			res := httptest.NewRecorder()
			svr.ServeHTTP(res, req)

			// Assert
			assertStatusCode(t, res, http.StatusNotFound)
		})
	})

	t.Run("Get request to /api/nearestPlaygrounds", func(t *testing.T) {
		t.Run("Returns a list of playgrounds ordered by proximity", func(t *testing.T) {
			req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?adress=42 Avenue de Flandre Paris")
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
			req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?test=42 Avenue de Flandre Paris")
			res := httptest.NewRecorder()

			svr.ServeHTTP(res, req)

			assertStatusCode(t, res, http.StatusBadRequest)
		})
		t.Run("Returns bad request if empty adress parameter", func(t *testing.T) {
			req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?adress=")
			res := httptest.NewRecorder()

			svr.ServeHTTP(res, req)

			assertStatusCode(t, res, http.StatusBadRequest)
		})
	})
}

func TestPostAPIs(t *testing.T) {
	t.Run("Post request to /api/playgrounds returns 400", func(t *testing.T) {
		// Arrange
		svr := server.New(nil, nil, nil, dummyMiddlewares)

		// Act
		req, err := http.NewRequest(http.MethodPost, "/api/playgrounds", nil)
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
