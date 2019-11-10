package test

import (
	"net/http"
	"reflect"
	"strings"
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

func NewPostFormRequest(t *testing.T, url string, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("Couldn't create request, %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func NewDeleteRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatalf("Couldn't create request, %v", err)
	}
	return req
}

func NewPutRequest(t *testing.T, url string, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("Couldn't create request, %v", err)
	}
	return req
}

func AssertPlayground(t *testing.T, got, want store.Playground) {
	t.Helper()
	if got.Name != want.Name || got.Address != want.Address || got.Lat != want.Lat {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func AssertPlaygrounds(t *testing.T, got, want store.Playgrounds) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func AssertComment(t *testing.T, got, want store.Comment) {
	t.Helper()
	if got.ID != want.ID || got.Author != want.Author || got.Content != want.Content {
		t.Errorf("Got %v, want %v", got, want)
	}
}
