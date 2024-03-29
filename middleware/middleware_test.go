package middleware

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yousseffarkhani/playground/backend2/authentication"
	"github.com/yousseffarkhani/playground/backend2/server"
)

type mockHandler struct {
	claims *authentication.Claims
	called bool
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.claims, _ = r.Context().Value("claims").(*authentication.Claims)
	m.called = true
}

func TestIsLogged(t *testing.T) {
	want := "test"
	mockHandler := &mockHandler{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Problem setting cookie jar, %s", err)
	}
	client := &http.Client{
		Jar: jar,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authentication.SetJwtCookie(w, want)
	})
	mux.Handle("/isLogged", isLogged(mockHandler))

	svr := httptest.NewServer(mux)
	defer svr.Close()

	// Setting JWT cookie
	_, err = client.Get(svr.URL + "/")
	if err != nil {
		t.Fatalf("Couldn't get a response, %s", err)
	}

	// Check if isLogged has set a context with claims in it
	_, err = client.Get(svr.URL + "/isLogged")
	if err != nil {
		t.Fatalf("Couldn't get a response, %s", err)
	}

	if mockHandler.claims == nil {
		t.Fatal("Should get claims from context")
	}

	got := mockHandler.claims.Username
	if got != want {
		t.Errorf("got : %q, want %q", got, want)
	}
}

func TestRefreshJWT(t *testing.T) {
	mockHandler := &mockHandler{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Problem setting cookie jar, %s", err)
	}
	client := &http.Client{
		Jar: jar,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authentication.SetJwtCookie(w, "test")
	})
	mux.Handle("/refreshJWT", isLogged(refreshJWT(mockHandler)))

	svr := httptest.NewServer(mux)
	defer svr.Close()

	// Setting JWT cookie
	_, err = client.Get(svr.URL + "/")
	if err != nil {
		t.Fatalf("Couldn't get a response, %s", err)
	}

	_, err = client.Get(svr.URL + "/refreshJWT")
	if err != nil {
		t.Fatalf("Couldn't get a response, %s", err)
	}

	if mockHandler.claims == nil {
		t.Fatal("Should get claims from context")
	}

	got := time.Unix(mockHandler.claims.ExpiresAt, 0)
	if got.Sub(time.Now()) < 5*time.Minute {
		t.Errorf("Refresh middleware didn't refresh JWT, got : %s", got.Format(time.RFC3339))
	}
}

func TestIsAuthorized(t *testing.T) {
	mockHandler := &mockHandler{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Problem setting cookie jar, %s", err)
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authentication.SetJwtCookie(w, "test")
	})
	mux.Handle("/authorized", isLogged(refreshJWT(isAuthorized(mockHandler))))

	svr := httptest.NewServer(mux)
	defer svr.Close()

	t.Run("Redirects to login if no JWT", func(t *testing.T) {
		resp, err := client.Get(svr.URL + "/authorized")
		if err != nil {
			t.Fatalf("Couldn't get a response, %s", err)
		}

		if len(resp.Header["Location"]) == 0 {
			t.Fatal("Header should have location")
		}

		got := resp.Header["Location"][0]
		want := server.URLLogin
		if got != want {
			t.Errorf("Got : %s, want : %s", got, want)
		}
	})

	t.Run("Handler called if JWT present", func(t *testing.T) {
		// Setting JWT cookie
		_, err = client.Get(svr.URL + "/")
		if err != nil {
			t.Errorf("Couldn't get a response, %s", err)
		}

		_, err = client.Get(svr.URL + "/authorized")
		if err != nil {
			t.Errorf("Couldn't get a response, %s", err)
		}

		if !mockHandler.called {
			t.Error("Handler should be called")
		}
	})
}
