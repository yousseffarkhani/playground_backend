package store

import (
	"errors"
	"math"
	"sort"

	"github.com/yousseffarkhani/playground/backend2/server"
)

type Playgrounds []server.Playground

var ErrorNotFoundPlayground = errors.New("Playground doesn't exist")

func (p Playgrounds) Find(ID int) (server.Playground, error) {
	if ID > len(p) || ID <= 0 {
		return server.Playground{}, ErrorNotFoundPlayground
	}
	return p[ID-1], nil
}
func (p Playgrounds) FindNearestPlaygrounds(client GeolocationClient, adress string) Playgrounds {
	long, lat := client.GetLongAndLat(adress)
	sort.Slice(p, func(i, j int) bool {
		distanceFromAdressToI := math.Abs(long-p[i].Long) + math.Abs(lat-p[i].Lat)
		distanceFromAdressToJ := math.Abs(long-p[j].Long) + math.Abs(lat-p[j].Lat)
		return distanceFromAdressToI > distanceFromAdressToJ
	})
	return p
}

type GeolocationClient interface {
	GetLongAndLat(adress string) (long, lat float64)
}
