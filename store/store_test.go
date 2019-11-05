package store_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/test"

	"github.com/yousseffarkhani/playground/backend2/store"
)

func TestPlaygroundDatabase(t *testing.T) {
	submittedPlaygroundStore := &store.SubmittedPlaygroundStore{}
	file, removeFile := createTempFile(t, `[
		{"Name": "aaaa", "Address": "aaaa"}]`)
	defer removeFile()
	str, _ := store.New(file)
	database := store.PlaygroundDatabase{
		MainPlaygroundStore:      str,
		SubmittedPlaygroundStore: submittedPlaygroundStore,
	}

	newPlayground1 := store.Playground{
		Name:       "bbbb",
		Address:    "bbbb",
		PostalCode: "75001",
		City:       "b",
		Department: "b",
		Author:     "Youssef",
	}
	newPlayground2 := store.Playground{
		Name:       "Cccc",
		Address:    "cccc",
		PostalCode: "75001",
		City:       "c",
		Department: "c",
		Author:     "Youssef",
	}
	newPlayground3 := store.Playground{
		Name:       "DdDD",
		Address:    "dddD",
		PostalCode: "75019",
		City:       "Paris",
		Department: "Paris",
		Author:     "Youssef",
	}
	t.Run("Playground returns an error if playground doesn't exist", func(t *testing.T) {
		_, got := str.Playground(2)

		assertError(t, got, store.ErrorNotFoundPlayground)
	})
	t.Run("Submit playground", func(t *testing.T) {
		cases := []store.Playground{
			newPlayground1,
			newPlayground2,
			newPlayground3,
		}
		for index, newPlayground := range cases {
			t.Run(" adds a playground to submitted playgrounds store and increments ID", func(t *testing.T) {
				errorsMap := database.SubmitPlayground(newPlayground)
				if len(errorsMap) > 0 {
					t.Fatalf("There shouldn't be an error, %s", errorsMap)
				}

				playgroundSubmitted, err := database.SubmittedPlaygroundStore.Playground(index + 1)
				if err != nil {
					t.Fatalf("There shouldn't be an error, %s", err)
				}

				test.AssertPlayground(t, playgroundSubmitted, newPlayground)
			})
		}

		t.Run(" returns an error if playground already exists in main store or submitted playgrounds store (same name / same address)", func(t *testing.T) {
			cases := map[string]store.Playground{
				"same name (Capitalized)": store.Playground{
					Name:       strings.ToUpper(newPlayground1.Name),
					Address:    "test",
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Author:     "Youssef",
				},
				"same address": store.Playground{
					Name:       "test",
					Address:    strings.ToUpper(newPlayground1.Address),
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Author:     "Youssef",
				},
				"Existing playground name in main store": store.Playground{
					Name:       "AaaA",
					Address:    "test",
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Author:     "Youssef",
				},
				"Existing playground address in main store": store.Playground{
					Name:       "test",
					Address:    "AAAA",
					PostalCode: "75019",
					City:       "Paris",
					Department: "Paris",
					Author:     "Youssef",
				},
			}
			for errorDescription, playground := range cases {
				err := database.SubmitPlayground(playground)
				if err == nil {
					t.Errorf("There should be an error, %q", errorDescription)
				}
			}
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
					Address:    "  ",
					PostalCode: "75001",
					City:       "Paris",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "   ",
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
					City:       "   ",
					Department: "Paris",
				},
				store.Playground{
					Name:       "test",
					Address:    "test",
					PostalCode: "75555",
					City:       "Paris",
					Department: "  ",
				},
			}
			for _, playground := range cases {
				err := database.SubmitPlayground(playground)

				if err == nil {
					t.Errorf("There should be an error : %+v", playground)
				}
			}
		})
	})
	t.Run("All playgrounds returns playgrounds sorted by name", func(t *testing.T) {
		got := database.SubmittedPlaygroundStore.AllPlaygrounds()
		newPlayground1.ID = 1
		newPlayground2.ID = 2
		newPlayground3.ID = 3
		want := []store.Playground{
			newPlayground1,
			newPlayground2,
			newPlayground3,
		}
		test.AssertPlaygrounds(t, got, want)
	})
	t.Run("DeletePlayground deletes playground from submitted playgrounds", func(t *testing.T) {
		deleteID := newPlayground2.ID
		originalLength := len(submittedPlaygroundStore.AllPlaygrounds())
		submittedPlaygroundStore.DeletePlayground(deleteID)

		postDeleteLength := len(submittedPlaygroundStore.AllPlaygrounds())
		if postDeleteLength >= originalLength {
			t.Fatalf("Slice length %d, should be lower than the original length %d", postDeleteLength, originalLength)
		}

		for _, playground := range submittedPlaygroundStore.AllPlaygrounds() {
			if playground.ID == deleteID {
				t.Errorf("Playground %s should have been deleted", playground.Name)
			}
		}
	})
	t.Run("Add Playground ", func(t *testing.T) {
		t.Run("ADDS a new playground, INCREMENTS ID and deletes entry from submittedPlaygrounds", func(t *testing.T) {
			newPlayground1.Long = 2
			newPlayground1.Lat = 2

			errorsMap := database.AddPlayground(newPlayground1, newPlayground1.ID)
			if len(errorsMap) > 0 {
				t.Fatalf("Couldn't add playground, %v", errorsMap)
			}

			got, err := str.Playground(2)
			if err != nil {
				t.Fatalf("There shouldn't be an error, %s", err)
			}

			test.AssertPlayground(t, got, newPlayground1)

			_, err = submittedPlaygroundStore.Playground(newPlayground1.ID)
			if err == nil {
				t.Errorf("There should be an error")
			}
		})
		t.Run("Returns an error ", func(t *testing.T) {
			newPlayground := store.Playground{
				Name:       "EEEE",
				Address:    "eeee",
				PostalCode: "75019",
				City:       "Paris",
				Department: "Paris",
				Long:       1,
				Lat:        1,
				Author:     "Youssef",
			}
			t.Run("if playgroundID doesn't match any submitted playground", func(t *testing.T) {
				errorsMap := database.AddPlayground(newPlayground, 2)
				if len(errorsMap) == 0 {
					t.Fatalf("There should be an error \n")
				}
			})
			t.Run("if playgroundID doesn't match playground name", func(t *testing.T) {
				errorsMap := database.AddPlayground(newPlayground, 3)
				if len(errorsMap) == 0 {
					t.Fatalf("There should be an error \n")
				}
			})
			t.Run("if playground already exists (same long and lat)", func(t *testing.T) {
				newPlayground3.Long = 2
				newPlayground3.Lat = 2

				err := database.AddPlayground(newPlayground3, newPlayground3.ID)
				if err == nil {
					t.Errorf("There should be an error")
				}
			})
		})
	})
}

func TestNew(t *testing.T) {
	t.Run("New WORKS with correct file", func(t *testing.T) {
		file, removeFile := createTempFile(t, `[
		{"Name": "b"},{"Name": "a"} ]`)
		defer removeFile()

		str, _ := store.New(file)
		playground1 := store.Playground{Name: "a", ID: 1}
		playground2 := store.Playground{Name: "b", ID: 2}

		t.Run("Playground RETURNS the right playground", func(t *testing.T) {
			got, _ := str.Playground(2)

			test.AssertPlayground(t, got, playground2)
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
	t.Run("New works even with an empty file", func(t *testing.T) {
		file, removeFile := createTempFile(t, "")
		defer removeFile()

		str, _ := store.New(file)

		got := str.AllPlaygrounds()

		test.AssertPlaygrounds(t, got, store.Playgrounds{})
	})
	t.Run("New returns an error if file isn't JSON formatted", func(t *testing.T) {
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
