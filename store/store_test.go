package store_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/test"

	"github.com/yousseffarkhani/playground/backend2/store"
)

func TestStore(t *testing.T) {
	t.Run("Store works with correct file", func(t *testing.T) {
		file, removeFile := createTempFile(t, `[
		{"Name": "b"},{"Name": "a"} ]`)
		defer removeFile()
		str, _ := store.New(file)
		playground1 := store.Playground{Name: "a", ID: 1}
		playground2 := store.Playground{Name: "b", ID: 2}
		t.Run("Playground returns the right playground", func(t *testing.T) {
			got, _ := str.Playground(2)

			test.AssertPlayground(t, got, playground2)
		})
		t.Run("Playground returns an error if playground doesn't exist", func(t *testing.T) {
			_, got := str.Playground(0)
			want := store.ErrorNotFoundPlayground

			assertError(t, got, want)

			_, got = str.Playground(3)
			want = store.ErrorNotFoundPlayground

			assertError(t, got, want)
		})
		t.Run("AllPlaygrounds returns playgrounds SORTED by name", func(t *testing.T) {
			got := str.AllPlaygrounds()
			want := store.Playgrounds{
				playground1,
				playground2,
			}

			test.AssertPlaygrounds(t, got, want)
		})
		t.Run("New adds IDs to playgrounds starting with 1 and ordered by name", func(t *testing.T) {
			playground, _ := str.Playground(1)
			got := playground.ID
			want := playground1.ID

			if got != want {
				t.Errorf("got : %d, want : %d", got, want)
			}
		})
	})

	t.Run("Store works even with an empty file", func(t *testing.T) {
		file, removeFile := createTempFile(t, "")
		defer removeFile()

		str, _ := store.New(file)

		got := str.AllPlaygrounds()

		test.AssertPlaygrounds(t, got, store.Playgrounds{})
	})
	t.Run("Store returns an error if file isn't JSON formatted", func(t *testing.T) {
		file, removeFile := createTempFile(t, "This is a test")
		defer removeFile()

		_, got := store.New(file)

		assertError(t, got, store.ErrorParsingJson)
	})
	t.Run("Add new playground", func(t *testing.T) {
		// Assert la casse pour vérifier qu'il n'existe pas déjà
		// Assert empty playground
	})
	t.Run("Returns an error if playground already exists", func(t *testing.T) {})
}

func assertError(t *testing.T, got, want error) {
	t.Helper()
	if got != want {
		t.Errorf("Got %q, want %q", got, want)
	}
}

func createTempFile(t *testing.T, data string) (*os.File, func()) {
	t.Helper()
	tempFile, err := ioutil.TempFile("", "testdb")
	if err != nil {
		t.Fatalf("Could't create temp file, %s", err)
	}

	removeTempFile := func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}
	tempFile.Write([]byte(data))
	return tempFile, removeTempFile
}
