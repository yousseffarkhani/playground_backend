package middleware_test

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/middleware"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

type mockHandler struct {
	claims *authentication.Claims
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.claims, _ = r.Context().Value("claims").(*authentication.Claims)
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
	mux.Handle("/isLogged", middleware.IsLogged(mockHandler))

	svr := httptest.NewServer(mux)
	defer svr.Close()

	// Setting JWT cookie
	_, err = client.Get(svr.URL + "/")
	if err != nil {
		t.Errorf("Couldn't get a response, %s", err)
	}

	// Check if isLogged has set a context with claims in it
	_, err = client.Get(svr.URL + "/isLogged")
	if err != nil {
		t.Errorf("Couldn't get a response, %s", err)
	}

	if mockHandler.claims == nil {
		t.Fatal("Should get claims from context")
	}

	got := mockHandler.claims.Username
	if got != want {
		t.Errorf("got : %q, want %q", got, want)
	}
}
