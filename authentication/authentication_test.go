package authentication_test

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

func TestSetParseAndUnsetJWT(t *testing.T) {
	want := "test"

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Problem setting cookie jar, %s", err)
	}
	client := &http.Client{
		Jar: jar,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		authentication.SetJwtCookie(w, want)
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		authentication.UnsetJWTCookie(w)
	})

	svr := httptest.NewServer(mux)
	defer svr.Close()

	resp, err := client.Get(svr.URL + "/login")
	if err != nil {
		t.Errorf("Couldn't get a response, %s", err)
	}
	defer resp.Body.Close()

	u, err := url.Parse(svr.URL)
	if err != nil {
		t.Fatalf("Couldn't parse url, %s", err)
	}
	cookies := client.Jar.Cookies(u)

	if len(cookies) != 1 {
		t.Error("Client should have cookie")
	}
	for _, c := range cookies {
		if c.Value == "" {
			t.Fatal("Cookie should contain JWT value")
		}
		claims, _, err := authentication.ParseCookie(c)
		if err != nil {
			t.Fatalf("Couldn't parse cookie, %s", err)
		}
		got := claims.Username
		if got != "test" {
			t.Errorf("Username is not correct, got : %s, want : %s", got, want)
		}
	}

	resp, err = client.Get(svr.URL + "/logout")
	if err != nil {
		t.Errorf("Couldn't get a response, %s", err)
	}
	defer resp.Body.Close()

	if len(client.Jar.Cookies(u)) != 0 {
		t.Error("Cookie value should be empty")
	}
}
