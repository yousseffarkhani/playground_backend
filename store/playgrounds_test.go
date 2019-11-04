package store_test

import (
	"reflect"
	"testing"

	"github.com/yousseffarkhani/playground/backend2/test"

	"github.com/yousseffarkhani/playground/backend2/store"
)

type stubClient struct{}

func (s stubClient) GetLongAndLat(address string) (long, lat float64, err error) {
	long = 2.372452
	lat = 48.886835
	return long, lat, nil
}
func TestPlaygrounds(t *testing.T) {
	t.Run("FindNearestPlaygrounds returns playgrounds from nearest to farthest", func(t *testing.T) {
		// {TEP JARDINS SAINT PAUL 2.36016 48.8532}
		// Distance : 3,37km Temps : 55 min A pied : 4,3km
		nearestPlayground := store.Playground{
			Name: "TEP JARDINS SAINT PAUL",
			Long: 2.36016000,
			Lat:  48.85320000,
		}
		// {ETABLISSEMENT FENELON 2.31718 48.87867}
		// Distance : 4,19km Temps : 59 min A pied : 4,5km
		intermediatePlayground := store.Playground{
			Name: "ETABLISSEMENT FENELON",
			Long: 2.31718,
			Lat:  48.87867,
		}
		// {LYCEE VICTOR DURUY 2.31565 48.8533}
		// Distance : 5,7km Temps : 1h24 A pied : 6,5km
		farthestPlayground := store.Playground{
			Name: "LYCEE VICTOR DURUY",
			Long: 2.31565,
			Lat:  48.8533,
		}
		playgrounds := store.Playgrounds{intermediatePlayground, farthestPlayground, nearestPlayground}
		client := stubClient{}

		want := store.Playgrounds{
			nearestPlayground,
			intermediatePlayground,
			farthestPlayground,
		}

		got, _ := playgrounds.FindNearestPlaygrounds(client, "42 avenue de Flandre Paris")

		test.AssertPlaygrounds(t, got, want)
	})
	t.Run("Find returns correct playground", func(t *testing.T) {
		playgrounds := setupPlaygrounds()
		want := playgrounds[0]

		got, _ := playgrounds.Find(1)

		test.AssertPlayground(t, got, want)
	})
	t.Run("Find returns error if playground doesn't exist", func(t *testing.T) {
		playgrounds := setupPlaygrounds()

		_, got := playgrounds.Find(0)
		assertError(t, got, store.ErrorNotFoundPlayground)

		_, got = playgrounds.Find(3)
		assertError(t, got, store.ErrorNotFoundPlayground)
	})
}

func TestComments(t *testing.T) {
	t.Run("AddComment adds a new comment with correct ID", func(t *testing.T) {
		playground := store.Playground{}
		cases := store.Comments{
			store.Comment{
				Author:  "test",
				Content: "test",
			},
			store.Comment{
				Author:  "test1",
				Content: "test1",
			},
		}
		for index, comment := range cases {
			err := playground.AddComment(comment)
			if err != nil {
				t.Fatalf("Couldn't add comment, %s", err)
			}
			got, err := playground.FindComment(index + 1)
			if err != nil {
				t.Fatalf("Couldn't get comment, %s", err)
			}
			comment.ID = index + 1
			want := comment

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got : %v, want : %v", got, want)
			}
		}
	})
	t.Run("AddComment returns an error if content or author empty", func(t *testing.T) {
		playground := store.Playground{}
		cases := store.Comments{
			store.Comment{
				Author:  "  ",
				Content: "test",
			},
			store.Comment{
				Author:  "test1",
				Content: "   ",
			},
		}
		for _, comment := range cases {
			err := playground.AddComment(comment)
			if err == nil {
				t.Fatalf("There should be an error")
			}
		}
	})
	// playgrounds := store.Playgrounds{intermediatePlayground, farthestPlayground, nearestPlayground}
	// client := stubClient{}

	// want := store.Playgrounds{
	// 	nearestPlayground,
	// 	intermediatePlayground,
	// 	farthestPlayground,
	// }

	// got, _ := playgrounds.FindNearestPlaygrounds(client, "42 avenue de Flandre Paris")

	// test.AssertPlaygrounds(t, got, want)
	// t.Run("Find returns correct playground", func(t *testing.T) {
	// 	playgrounds := setupPlaygrounds()
	// 	want := playgrounds[0]

	// 	got, _ := playgrounds.Find(1)

	// 	test.AssertPlayground(t, got, want)
	// })
	// t.Run("Find returns error if playground doesn't exist", func(t *testing.T) {
	// 	playgrounds := setupPlaygrounds()

	// 	_, got := playgrounds.Find(0)
	// 	assertError(t, got, store.ErrorNotFoundPlayground)

	// 	_, got = playgrounds.Find(3)
	// 	assertError(t, got, store.ErrorNotFoundPlayground)
	// })
}

func setupPlaygrounds() store.Playgrounds {
	playground1 := store.Playground{
		Name: "1",
		ID:   1,
	}
	playground2 := store.Playground{
		Name: "2",
		ID:   2,
	}
	playgrounds := store.Playgrounds{playground1, playground2}
	return playgrounds
}
