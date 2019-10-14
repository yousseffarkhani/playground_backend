package test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/store"
)

func NewGetRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Couldn't create request, %v", err)
	}
	return req
}

func AssertPlayground(t *testing.T, got, want store.Playground) {
	t.Helper()
	if got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func AssertPlaygrounds(t *testing.T, got, want store.Playgrounds) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}
