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
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	PostalCode string  `json:"postal_code"`
	City       string  `json:"city"`
	Department string  `json:"department"`
	Long       float64 `json:"long"`
	Lat        float64 `json:"lat"`
	Coating    string  `json:"coating"`
	Type       string  `json:"type"`
	Open       bool    `json:"open"`
	ID         int
	Comments   Comments
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
var ErrorNotFoundComment = errors.New("Comment doesn't exist")

func (p Playgrounds) Find(ID int) (Playground, error) {
	for _, playground := range p {
		if playground.ID == ID {
			return playground, nil
		}
	}
	return Playground{}, ErrorNotFoundPlayground
}
func (p Playgrounds) FindNearestPlaygrounds(client GeolocationClient, address string) (Playgrounds, error) {
	long, lat, err := client.GetLongAndLat(address)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get longitude and lattitude, %s", err)
	}
	playgroundsSorted := p.sortByProximity(long, lat)
	return playgroundsSorted, nil
}

type GeolocationClient interface {
	GetLongAndLat(address string) (long, lat float64, err error)
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
		distanceFromAddressToI := playgroundsSorted[i].calculateSquaredDistanceFrom(long, lat)
		distanceFromAddressToJ := playgroundsSorted[j].calculateSquaredDistanceFrom(long, lat)
		return distanceFromAddressToI < distanceFromAddressToJ
	})
	return playgroundsSorted
}

func (p Playground) calculateSquaredDistanceFrom(long, lat float64) float64 {
	return math.Pow(long-p.Long, 2) + math.Pow(lat-p.Lat, 2)
}

func (p Playground) FindComment(commentID int) (Comment, error) {
	for _, comment := range p.Comments {
		if comment.ID == commentID {
			return comment, nil
		}
	}
	return Comment{}, ErrorNotFoundComment
}
