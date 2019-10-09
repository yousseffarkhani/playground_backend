package store_test

import (
	"testing"

	"github.com/yousseffarkhani/playground/backend2/server"

	"github.com/yousseffarkhani/playground/backend2/test"

	"github.com/yousseffarkhani/playground/backend2/store"
)

type stubClient struct {}

func (s stubClient) GetLongAndLat(adress string) (long, lat float64) {
	long = 2.333333
	lat = 48.866667
	return long, lat
}
func TestPlaygrounds(t *testing.T) {
	t.Run("FindNearestPlaygrounds returns playgrounds from nearest to farthest", func(t *testing.T) {
		nearestPlayground := server.Playground{
			Name: "TEP Neuve Saint pierre",
			Long: 2.36295000,
			Lat:  48.85351000,
		}
		intermediatePlayground := server.Playground{
			Name: "TEP neuve st paul",
			Long: 2.36016000,
			Lat:  48.85320000,
		}
		farthestPlayground := server.Playground{
			Name: "LYCEE DE LA ROCHEFOUCAULT",
			Long: 2.30541000,
			Lat:  48.86020000,
		}
		playgrounds := store.Playgrounds{intermediatePlayground, farthestPlayground, nearestPlayground}
		client := stubClient()

		want := store.Playgrounds{
			nearestPlayground,
			intermediatePlayground,
			farthestPlayground,
		}

		got := playgrounds.FindNearestPlaygrounds(client, "42 avenue de Flandre Paris")

		test.AssertPlaygrounds(t, got, want)
	})
	t.Run("Find returns correct playground", func(t *testing.T) {
		playground1 := server.Playground{
			Name: "1",
		}
		playground2 := server.Playground{
			Name: "2",
		}
		playgrounds := store.Playgrounds{playground1, playground2}
		want := playground1

		got, _ := playgrounds.Find(1)

		test.AssertPlayground(t, got, want)
	})
	t.Run("Find returns error if playground doesn't exist", func(t *testing.T) {
		playground1 := server.Playground{
			Name: "1",
		}
		playground2 := server.Playground{
			Name: "2",
		}
		playgrounds := store.Playgrounds{playground1, playground2}

		_, got := playgrounds.Find(0)
		assertError(t, got, store.ErrorNotFoundPlayground)

		_, got = playgrounds.Find(3)
		assertError(t, got, store.ErrorNotFoundPlayground)
	})
}
