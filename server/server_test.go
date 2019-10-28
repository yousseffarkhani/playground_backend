package server_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/configuration"
	"github.com/yousseffarkhani/playground/backend2/server"

	"github.com/yousseffarkhani/playground/backend2/store"
	"github.com/yousseffarkhani/playground/backend2/test"
)

var comment1 = store.Comment{
	ID:      1,
	Content: "Great Playground !",
	Author:  "Youssef",
}

var comment2 = store.Comment{
	ID:      2,
	Content: "Bad Playground !",
	Author:  "Clélia",
}
var comment3 = store.Comment{
	ID:      1,
	Content: "Ok Playground !",
	Author:  "Thibaut",
}

var playground1 = store.Playground{
	Name: "test1",
	Long: 2.36016000,
	Lat:  48.85320000,
	Comments: store.Comments{
		comment1,
		comment2,
	},
}
var playground2 = store.Playground{
	Name: "test2",
	Long: 2.31565,
	Lat:  48.8533,
	Comments: store.Comments{
		comment3,
	},
}
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

func (m *mockGeolocationClient) GetLongAndLat(address string) (long, lat float64, err error) {
	return 2.372452, 48.886835, nil
}

func TestAPIs(t *testing.T) {
	// Arrange
	str := &mockPlaygroundStore{playgrounds: playgrounds}
	client := &mockGeolocationClient{}

	svr := server.New(str, client, nil, dummyMiddlewares)

	t.Run("Playground APIs : ", func(t *testing.T) {
		t.Run(server.APIPlaygrounds, func(t *testing.T) {
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
		})
		t.Run(server.APIPlayground, func(t *testing.T) {
			t.Run("Get request to /api/playgrounds/1", func(t *testing.T) {
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
					req := test.NewGetRequest(t, server.APIPlaygrounds+"/1000")
					res := httptest.NewRecorder()
					svr.ServeHTTP(res, req)

					// Assert
					assertStatusCode(t, res, http.StatusNotFound)
				})
			})
		})
		t.Run(server.APINearestPlaygrounds, func(t *testing.T) {
			t.Run("Get request to /api/nearestPlaygrounds", func(t *testing.T) {
				t.Run("Returns a list of playgrounds ordered by proximity", func(t *testing.T) {
					req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?address=42 Avenue de Flandre Paris")
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
				t.Run("Returns bad request if no address parameter in query", func(t *testing.T) {
					req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?test=42 Avenue de Flandre Paris")
					res := httptest.NewRecorder()

					svr.ServeHTTP(res, req)

					assertStatusCode(t, res, http.StatusBadRequest)
				})
				t.Run("Returns bad request if empty address parameter", func(t *testing.T) {
					req := test.NewGetRequest(t, server.APINearestPlaygrounds+"?address=")
					res := httptest.NewRecorder()

					svr.ServeHTTP(res, req)

					assertStatusCode(t, res, http.StatusBadRequest)
				})
			})
		})
	})
	t.Run("Comments APIs : ", func(t *testing.T) {
		t.Run(server.APIComments, func(t *testing.T) {
			cases := map[string]store.Comments{
				"/api/playgrounds/1/comments": store.Comments{comment1, comment2},
				"/api/playgrounds/2/comments": store.Comments{comment3},
			}
			for URL, want := range cases {
				t.Run(fmt.Sprintf("Get request to %s", URL), func(t *testing.T) {
					// Act
					req := test.NewGetRequest(t, URL)
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
					t.Run("returns a list of comments", func(t *testing.T) {
						var got store.Comments
						err := json.NewDecoder(res.Body).Decode(&got)
						if err != nil {
							t.Fatalf("Unable to parse input %q into slice, '%v'", res.Body, err)
						}
						if !reflect.DeepEqual(got, want) {
							t.Errorf("Got %v, want %v", got, want)
						}
					})
				})
			}
			t.Run("Returns 404 if playground doesn't exist", func(t *testing.T) {
				// Act
				req := test.NewGetRequest(t, "/api/playgrounds/1000/comments")
				res := httptest.NewRecorder()
				svr.ServeHTTP(res, req)

				// Assert
				assertStatusCode(t, res, http.StatusNotFound)
			})
		})
		t.Run(server.APIComment, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				cases := map[string]store.Comment{
					"/api/playgrounds/1/comments/1": comment1,
					"/api/playgrounds/1/comments/2": comment2,
				}
				for URL, want := range cases {
					t.Run(fmt.Sprintf(" request to %s", URL), func(t *testing.T) {
						// Act
						req := test.NewGetRequest(t, URL)
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
						t.Run("returns a comment", func(t *testing.T) {
							var got store.Comment
							err := json.NewDecoder(res.Body).Decode(&got)
							if err != nil {
								t.Fatalf("Unable to parse input %q into comment, '%v'", res.Body, err)
							}
							if !reflect.DeepEqual(got, want) {
								t.Errorf("Got %v, want %v", got, want)
							}
						})
					})
				}

				t.Run("Returns 404 if playground or comment doesn't exist", func(t *testing.T) {
					// Act
					cases := []string{
						"/api/playgrounds/1000/comments/1",
						"/api/playgrounds/1/comments/1000",
					}
					for _, URL := range cases {
						req := test.NewGetRequest(t, URL)
						res := httptest.NewRecorder()
						svr.ServeHTTP(res, req)

						// Assert
						assertStatusCode(t, res, http.StatusNotFound)
					}
				})
				t.Run("Returns internal server error if query parameters are invalid", func(t *testing.T) {
					// Act
					cases := []string{
						"/api/playgrounds/aa/comments/1",
						"/api/playgrounds/1/comments/aaa",
					}
					for _, URL := range cases {
						req := test.NewGetRequest(t, URL)
						res := httptest.NewRecorder()
						svr.ServeHTTP(res, req)

						// Assert
						assertStatusCode(t, res, http.StatusInternalServerError)
					}
				})
			})
			/* t.Run("Post", func(t *testing.T) {
				t.Run(" records a new comment", func(t *testing.T) {
					req := test.NewPostFormRequest(t, "/api/playgrounds/1/comments", "comment=This is a nice playground")
					res := httptest.NewRecorder()

					svr.ServeHTTP(res, req)

					assertStatusCode(t, res, http.StatusAccepted)
					// assertContent
					// asset add comment called
					// Assert empty comment
					// assert author
					// Assert ID
					// Input : HTML Form comment
					// Decode input
					// Add comment
					// Add authorization middleware
				})
				t.Run(" returns status not found if playground doesn't exist", func(t *testing.T) {
					req := test.NewPostFormRequest(t, "/api/playgrounds/1000/comments", "comment=This is a nice playground")
					res := httptest.NewRecorder()
					svr.ServeHTTP(res, req)

					assertStatusCode(t, res, http.StatusNotFound)
				})
			}) */
			/* 		t.Run("Delete", func(t *testing.T) {
				t.Run(" returns status accepted", func(t *testing.T) {
					req := test.NewGetRequest(t, "/api/playgrounds/1/comments/1/delete")
					res := httptest.NewRecorder()
					svr.ServeHTTP(res, req)

					assertStatusCode(t, res, http.StatusAccepted)
				})
				t.Run(" returns status bad request (playground or comment doesn't exist)", func(t *testing.T) {
					cases := []string{
						"/api/playgrounds/1000/comments/1/delete",
						"/api/playgrounds/1/comments/1000/delete",
					}
					for _, c := range cases {
						req := test.NewGetRequest(t, c)
						res := httptest.NewRecorder()
						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusBadRequest)
					}
				})
			})
			t.Run("Update", func(t *testing.T) {
				t.Run(" returns status accepted", func(t *testing.T) {
					req := test.NewGetRequest(t, "/api/playgrounds/1/comments/1/update")
					res := httptest.NewRecorder()
					svr.ServeHTTP(res, req)

					assertStatusCode(t, res, http.StatusAccepted)
				})
				t.Run(" returns status bad request (playground or comment doesn't exist)", func(t *testing.T) {
					cases := []string{
						"/api/playgrounds/1000/comments/1/update",
						"/api/playgrounds/1/comments/1000/update",
					}
					for _, c := range cases {
						req := test.NewGetRequest(t, c)
						res := httptest.NewRecorder()
						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusBadRequest)
					}
				})
			}) */
		})
	})
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
	configuration.LoadEnvVariables()
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
			mockView: mockHomeView,
			expectedData: server.RenderingData{
				Username: "",
				Data:     nil},
		},
		server.URLPlaygrounds: testStruct{
			mockView: mockPlaygroundsView,
			expectedData: server.RenderingData{
				Username: "",
				Data:     playgrounds},
		},
		server.URLPlaygrounds + "/1": testStruct{
			mockView: mockPlaygroundView,
			expectedData: server.RenderingData{
				Username: "",
				Data:     playground1},
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

type mockMiddleware struct {
	called bool
}

func (m *mockMiddleware) ThenFunc(finalPage func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.called = true
		finalPage(w, r)
	})
}
func TestMiddlewares(t *testing.T) {
	configuration.LoadEnvVariables()
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
