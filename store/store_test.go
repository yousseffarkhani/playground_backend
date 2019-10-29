package store_test

import (
	"io/ioutil"
	"os"
	"strings"
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
		t.Run("Add new playground and increments ID", func(t *testing.T) {
			want := store.Playground{
				Name:       "c",
				Address:    "c",
				PostalCode: "75001",
				City:       "Paris",
				Department: "Paris",
				Long:       2,
				Lat:        2,
			}
			errorsMap := str.NewPlayground(want)
			if len(errorsMap) > 0 {
				t.Fatalf("Couldn't add playground, %v", errorsMap)
			}

			got, err := str.Playground(3)
			if err != nil {
				t.Fatalf("There shouldn't be an error, %s", err)
			}

			test.AssertPlayground(t, got, want)

		})
		t.Run("Returns an error if a field is empty and postal code is not a number or less than 5 numbers", func(t *testing.T) {
			cases := store.Playgrounds{
				store.Playground{
					Name:       " ",
					Address:    "test",
					PostalCode: "75001",
					City:       "Paris",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "",
					PostalCode: "75001",
					City:       "Paris",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "",
					City:       "Paris",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "aaaa",
					City:       "Paris",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "7555",
					City:       "Paris",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "75555",
					City:       "",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "75555",
					City:       "Paris",
					Department: "",
				},
			}
			for _, playground := range cases {
				err := str.NewPlayground(playground)

				if err == nil {
					t.Errorf("There should be an error : %+v", playground)
				}
			}
		})
		t.Run("Returns an error if playground already exists (same name / same address / same long and lat)", func(t *testing.T) {
			playground3 := store.Playground{
				Name:       "Gymnase de test",
				Address:    "2 avenue de flandre",
				PostalCode: "75019",
				City:       "Paris",
				Department: "Paris",
				Long:       5,
				Lat:        5,
			}
			errorsMap := str.NewPlayground(playground3)
			if len(errorsMap) > 0 {
				t.Fatalf("Couldn't add playground, %v", errorsMap)
			}

			cases := map[string]store.Playground{
				"same name (Capitalized)": store.Playground{
					Name:       strings.ToUpper(playground3.Name),
					Address:    "test",
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Long:       1,
					Lat:        1,
				},
				"same address": store.Playground{
					Name:       "test",
					Address:    strings.ToUpper(playground3.Address),
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Long:       1,
					Lat:        1,
				},
				"same long and lat": store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Long:       playground3.Long,
					Lat:        playground3.Lat,
				},
			}
			for errorDescription, playground := range cases {
				err := str.NewPlayground(playground)
				if err == nil {
					t.Errorf("There should be an error, %q", errorDescription)
				}
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
