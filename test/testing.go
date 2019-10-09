package test

import (
	"reflect"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/server"
)

func AssertPlayground(t *testing.T, got, want server.Playground) {
	t.Helper()
	if got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func AssertPlaygrounds(t *testing.T, got, want []server.Playground) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}
