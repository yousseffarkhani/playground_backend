package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
)

type Playground struct {
	Name           string  `json:"name"`
	Long           float64 `json:"long"`
	Lat            float64 `json:"lat"`
	Adress         string  `json:"adress"`
	Arrondissement int     `json:"arrondissement"`
	Dimensions     float64 `json:"dimensions"`
	Coating        string  `json:"coating"`
	Open           bool    `json:"open"`
	Lightning      bool    `json:"lightning"`
	ID             int
}

type Playgrounds []Playground

func NewPlaygrounds(input io.Reader) (Playgrounds, error) {
	var playgrounds Playgrounds
	err := json.NewDecoder(input).Decode(&playgrounds)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse input %q into slice, '%v'", input, err)
	}
	return playgrounds, nil
}

var ErrorNotFoundPlayground = errors.New("Playground doesn't exist")

func (p Playgrounds) Find(ID int) (Playground, error) {
	for _, playground := range p {
		if playground.ID == ID {
			return playground, nil
		}
	}
	return Playground{}, ErrorNotFoundPlayground
}
func (p Playgrounds) FindNearestPlaygrounds(client GeolocationClient, adress string) (Playgrounds, error) {
	long, lat, err := client.GetLongAndLat(adress)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get longitude and lattitude, %s", err)
	}
	playgroundsSorted := p.sortByProximity(long, lat)
	return playgroundsSorted, nil
}

type GeolocationClient interface {
	GetLongAndLat(adress string) (long, lat float64, err error)
}

func (p Playgrounds) sortByName() {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Name < p[j].Name
	})
}

func (p Playgrounds) sortByProximity(long, lat float64) Playgrounds {
	playgroundsSorted := make(Playgrounds, len(p))
	copy(playgroundsSorted, p)
	sort.SliceStable(playgroundsSorted, func(i, j int) bool {
		distanceFromAdressToI := playgroundsSorted[i].calculateSquaredDistanceFrom(long, lat)
		distanceFromAdressToJ := playgroundsSorted[j].calculateSquaredDistanceFrom(long, lat)
		return distanceFromAdressToI < distanceFromAdressToJ
	})
	return playgroundsSorted
}

func (p Playground) calculateSquaredDistanceFrom(long, lat float64) float64 {
	return math.Pow(long-p.Long, 2) + math.Pow(lat-p.Lat, 2)
}
