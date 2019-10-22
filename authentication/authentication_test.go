package authentication_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

func TestSetAndParseJWT(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authentication.SetJwtCookie(w, "test")
	})

	svr := httptest.NewServer(mux)
	defer svr.Close()

	resp, err := http.Get(svr.URL + "/")
	if err != nil {
		t.Errorf("Couldn't get a response, %s", err)
	}
	defer resp.Body.Close()

	for _, c := range resp.Cookies() {
		if c.Value != "" {
			claims, token, err := authentication.ParseCookie(c)
			if err != nil {
				t.Fatalf("Couldn't parse cookie, %s", err)
			}
			if !token.Valid {
				t.Errorf("Cookie is not valid")
			}
			if claims.Username != "test" {
				t.Errorf("Username is not correct, got : %s, want : %s", claims.Username, "test")
			}
		}
	}
}
