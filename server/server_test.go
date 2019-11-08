package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/yousseffarkhani/playground/backend2/authentication"
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
	Name:    "test1",
	Address: "42 avenue de Flandre",
	Long:    2.36016000,
	Lat:     48.85320000,
	Comments: store.Comments{
		comment1,
		comment2,
	},
}
var playground2 = store.Playground{
	Name:    "test2",
	Address: "43 avenue de Flandre",
	Long:    2.31565,
	Lat:     48.8533,
	Comments: store.Comments{
		comment3,
	},
}
var playgrounds = store.Playgrounds{
	playground1,
	playground2,
}
var dummyMiddlewares = map[string]server.Middleware{
	"isLogged":   &mockMiddleware{},
	"refresh":    &mockMiddleware{},
	"authorized": &mockMiddleware{},
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

func (m *mockPlaygroundStore) NewPlayground(newPlayground store.Playground) {
	m.playgrounds = append(m.playgrounds, newPlayground)
}

func (m *mockPlaygroundStore) DeletePlayground(ID int) {
}

type mockGeolocationClient struct{}

func (m *mockGeolocationClient) GetLongAndLat(address string) (long, lat float64, err error) {
	return 2.372452, 48.886835, nil
}

func TestAPIs(t *testing.T) {
	// Arrange
	playground1.ID = 1
	playground2.ID = 2
	var playgrounds = store.Playgrounds{
		playground1,
		playground2,
	}
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
						got, err := store.NewPlaygroundsFromJSON(res.Body)
						if err != nil {
							t.Fatalf("Unable to parse response into slice, '%v'", err)
						}

						test.AssertPlaygrounds(t, got, playgrounds)
					})
				})
			}
		})
		t.Run(server.APIPlayground, func(t *testing.T) {
			t.Run("GET", func(t *testing.T) {
				// Act
				cases := map[string]store.Playground{
					server.APIPlaygrounds + "/1": playground1,
					server.APIPlaygrounds + "/2": playground2,
				}
				for URL, want := range cases {
					t.Run(fmt.Sprintf(" request to %s", URL), func(t *testing.T) {
						// Act
						req := test.NewGetRequest(t, URL)
						res := httptest.NewRecorder()
						svr.ServeHTTP(res, req)

						// Assert
						t.Run(" returns status OK", func(t *testing.T) {
							assertStatusCode(t, res, http.StatusOK)
						})
						t.Run(" returns a JSON Content-Type ", func(t *testing.T) {
							assertHeader(t, res, "Content-Type", server.JsonContentType)
						})
						t.Run(" returns gzip Accept-Encoding", func(t *testing.T) {
							assertHeader(t, res, "Accept-Encoding", server.GzipAcceptEncoding)
						})
						t.Run(" returns a playground", func(t *testing.T) {
							var got store.Playground
							err := json.NewDecoder(res.Body).Decode(&got)
							if err != nil {
								t.Fatalf("Unable to parse input %q into comment, '%v'", res.Body, err)
							}
							test.AssertPlayground(t, got, want)
						})
					})
				}
				t.Run("Returns 404 if playground doesn't exist", func(t *testing.T) {
					// Act
					req := test.NewGetRequest(t, server.APIPlaygrounds+"/1000")
					res := httptest.NewRecorder()
					svr.ServeHTTP(res, req)

					// Assert
					assertStatusCode(t, res, http.StatusNotFound)
				})
			})
			t.Run("POST to ", func(t *testing.T) {
				t.Run(server.APISubmittedPlaygrounds, func(t *testing.T) {
					t.Run(" adds a new playground (with white spaces trimmed) to the submitted playground store", func(t *testing.T) {
						want := store.Playground{
							Name:       "test3",
							Address:    "44 avenue de Flandre",
							PostalCode: "75019",
							City:       "Paris",
							Department: "Paris",
						}
						mockForm := fmt.Sprintf("name= %s &address= %s &postal_code= %s &city= %s &department= %s ", want.Name, want.Address, want.PostalCode, want.City, want.Department)
						req := test.NewPostFormRequest(t, server.APISubmittedPlaygrounds, mockForm)
						req = setupRequestContext(req)
						res := httptest.NewRecorder()

						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusAccepted)

						req = test.NewGetRequest(t, server.APISubmittedPlaygrounds)
						res = httptest.NewRecorder()

						svr.ServeHTTP(res, req)

						got, err := store.NewPlaygroundsFromJSON(res.Body)
						if err != nil {
							t.Fatalf("Unable to parse response into slice, '%v'", err)
						}
						if len(got) == 0 {
							t.Fatalf("Response is empty")
						}
						test.AssertPlayground(t, got[0], want)
					})
					t.Run(" returns bad request", func(t *testing.T) {
						cases := map[string]store.Playground{
							" if there is an empty form value": store.Playground{
								Name:       "test4",
								Address:    "test",
								PostalCode: "  ",
								City:       "Paris",
								Department: "Paris",
							},
							" if there is already the same name in submitted playgrounds": store.Playground{
								Name:       "TeSt3",
								Address:    "test",
								PostalCode: "75019",
								City:       "Paris",
								Department: "Paris",
							},
							" if there is already the same address in submitted playgrounds": store.Playground{
								Name:       "test4",
								Address:    "44 AVENUE de Flandre",
								PostalCode: "75019",
								City:       "Paris",
								Department: "Paris",
							},
							" if there is already the same name in main playgrounds": store.Playground{
								Name:       "TeSt2",
								Address:    "test",
								PostalCode: "75019",
								City:       "Paris",
								Department: "Paris",
							},
							" if there is already the same address in main playgrounds": store.Playground{
								Name:       "test4",
								Address:    "42 AVENUE de Flandre",
								PostalCode: "75019",
								City:       "Paris",
								Department: "Paris",
							},
						}

						for description, playground := range cases {
							t.Run(description, func(t *testing.T) {
								mockForm := fmt.Sprintf("name= %s &address= %s &postal_code= %s &city= %s &department= %s ", playground.Name, playground.Address, playground.PostalCode, playground.City, playground.Department)
								req := test.NewPostFormRequest(t, server.APISubmittedPlaygrounds, mockForm)
								req = setupRequestContext(req)
								res := httptest.NewRecorder()

								svr.ServeHTTP(res, req)

								assertStatusCode(t, res, http.StatusBadRequest)
							})
						}
					})
				})

				t.Run(server.APISubmittedPlayground, func(t *testing.T) {
					// TODO
					// 1. Test to see if p.database.SubmittedPlaygroundStore.DeletePlayground() has been called
					// 2. Test to see if it returns a StatusInternalServerError

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

						got, err := store.NewPlaygroundsFromJSON(res.Body)
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
				t.Run("Post", func(t *testing.T) {
					t.Run(" records a new comment trimmed and increments ID", func(t *testing.T) {
						want := store.Comment{
							Author:  "Youssef",
							Content: "   This is a nice playground",
							ID:      3,
						}
						req := test.NewPostFormRequest(t, "/api/playgrounds/1/comments", fmt.Sprintf("comment=  %s", want.Content))
						req = setupRequestContext(req)
						res := httptest.NewRecorder()

						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusAccepted)

						req = test.NewGetRequest(t, "/api/playgrounds/1/comments/3")
						res = httptest.NewRecorder()
						svr.ServeHTTP(res, req)
						var got store.Comment
						json.NewDecoder(res.Body).Decode(&got)
						if reflect.DeepEqual(got, want) {
							t.Errorf("got : %+v, want : %+v", got, want)
						}
					})
					t.Run(" returns a status bad request if comment is empty", func(t *testing.T) {
						req := test.NewPostFormRequest(t, "/api/playgrounds/1/comments", "comment=   ")
						req = setupRequestContext(req)
						res := httptest.NewRecorder()

						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusBadRequest)
					})
					t.Run(" returns status not found if playground doesn't exist", func(t *testing.T) {
						req := test.NewPostFormRequest(t, "/api/playgrounds/1000/comments", "comment=This is a nice playground")
						req = setupRequestContext(req)
						res := httptest.NewRecorder()

						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusBadRequest)
					})
				})
				t.Run("Delete", func(t *testing.T) {
					t.Run(" returns status accepted and deletes comment if request comes from the author", func(t *testing.T) {
						req := test.NewDeleteRequest(t, "/api/playgrounds/1/comments/1")
						req = setupRequestContext(req)
						res := httptest.NewRecorder()
						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusAccepted)

						req = test.NewGetRequest(t, "/api/playgrounds/1/comments")
						res = httptest.NewRecorder()

						svr.ServeHTTP(res, req)

						var got store.Comments
						err := json.NewDecoder(res.Body).Decode(&got)
						if err != nil {
							t.Fatalf("Unable to parse input into comment, '%v'", err)
						}
						for _, comment := range got {
							if comment.ID == 1 {
								t.Errorf("This comment should be deleted")
							}
						}
					})
					t.Run(" returns status bad request (playground or comment doesn't exist)", func(t *testing.T) {
						cases := []string{
							"/api/playgrounds/1000/comments/1",
							"/api/playgrounds/1/comments/1000",
						}
						for _, URL := range cases {
							req := test.NewDeleteRequest(t, URL)
							res := httptest.NewRecorder()
							req = setupRequestContext(req)

							svr.ServeHTTP(res, req)

							assertStatusCode(t, res, http.StatusBadRequest)
						}
					})
					t.Run(" returns status bad request if comment doesn't belong to requester", func(t *testing.T) {
						req := test.NewDeleteRequest(t, "/api/playgrounds/1/comments/2")
						res := httptest.NewRecorder()
						req = setupRequestContext(req)

						svr.ServeHTTP(res, req)

						assertStatusCode(t, res, http.StatusBadRequest)
					})
				})
				/*	t.Run("Update", func(t *testing.T) {
					t.Run(" returns status accepted and updates comment if request comes from the author", func(t *testing.T) {
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
	})
}

func (m *mockPlaygroundStore) AddComment(playgroundID int, newComment store.Comment) error {
	_, index, err := m.playgrounds.Find(playgroundID)
	if err != nil {
		return err
	}
	err = m.playgrounds[index].AddComment(newComment)
	if err != nil {
		return err
	}
	return nil
}

func (m *mockPlaygroundStore) DeleteComment(playgroundID, commentID int, username string) error {
	playground, index, err := m.playgrounds.Find(playgroundID)
	if err != nil {
		return err
	}
	comment, err := playground.FindComment(commentID)
	if err != nil {
		return err
	}
	if !comment.IsAuthor(username) {
		return errors.New("Requester is not the author")
	}
	err = m.playgrounds[index].DeleteComment(commentID)
	if err != nil {
		return err
	}
	return nil
}

func setupRequestContext(req *http.Request) *http.Request {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "claims", &authentication.Claims{
		Username: "Youssef",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		},
	})
	return req.WithContext(ctx)
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
	playground1.ID = 0
	playground2.ID = 0
	var playgrounds = store.Playgrounds{
		playground1,
		playground2,
	}
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
	mockRefreshJWT := &mockMiddleware{}
	mockIsAuthorized := &mockMiddleware{}
	middlewares := map[string]server.Middleware{
		"isLogged":   mockIsLogged,
		"refresh":    mockRefreshJWT,
		"authorized": mockIsAuthorized,
	}
	str := &mockPlaygroundStore{}

	svr := server.New(str, nil, nil, middlewares)

	t.Run(fmt.Sprintf("isLogged middleware is called on route %q", server.URLLogin), func(t *testing.T) {
		req := test.NewGetRequest(t, server.URLLogin)

		svr.ServeHTTP(httptest.NewRecorder(), req)

		if mockIsLogged.called != true {
			t.Errorf("IsLogged middleware hasn't been called")
		}
		mockIsLogged.called = false
	})

	cases := []string{
		server.URLHome,
		server.URLPlaygrounds,
		server.URLPlaygrounds + "/1",
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("refresh middleware is called on route %q", c), func(t *testing.T) {
			req := test.NewGetRequest(t, c)

			svr.ServeHTTP(httptest.NewRecorder(), req)

			if mockRefreshJWT.called != true {
				t.Errorf("RefreshJWT middleware hasn't been called")
			}
			mockRefreshJWT.called = false
		})
	}
	tests := map[string]string{
		server.URLSubmitPlayground:            "GET",
		server.URLSubmittedPlaygrounds:        "GET",
		server.URLSubmittedPlaygrounds + "/1": "GET",
		server.APISubmittedPlaygrounds:        "POST",
		server.APIPlaygrounds:                 "POST",
		server.APISubmittedPlayground:         "POST",
		server.APIComments:                    "POST",
	}
	for url, method := range tests {
		t.Run(fmt.Sprintf("Authorized middleware is called on route %q", url), func(t *testing.T) {
			var req *http.Request
			if method == "GET" {
				req = test.NewGetRequest(t, url)
			} else {
				req = test.NewPostFormRequest(t, url, "")
			}

			svr.ServeHTTP(httptest.NewRecorder(), req)

			if mockIsAuthorized.called != true {
				t.Errorf("IsAuthorized middleware hasn't been called")
			}
			mockIsAuthorized.called = false
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
