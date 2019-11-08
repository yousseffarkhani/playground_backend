package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"
)

var ErrorNotFoundPlayground = errors.New("Playground doesn't exist")
var ErrorNotFoundComment = errors.New("Comment doesn't exist")

type Playground struct {
	Name             string  `json:"name"`
	Address          string  `json:"address"`
	PostalCode       string  `json:"postal_code"`
	City             string  `json:"city"`
	Department       string  `json:"department"`
	Long             float64 `json:"long"`
	Lat              float64 `json:"lat"`
	Coating          string  `json:"coating"`
	Type             string  `json:"type"`
	Open             bool    `json:"open"`
	ID               int
	Author           string
	TimeOfSubmission time.Time
	Comments         Comments
}

type Playgrounds []Playground

type Comment struct {
	ID int
	// PlaygroundID int
	Content          string
	Author           string
	TimeOfSubmission time.Time
}

type Comments []Comment

func NewPlaygroundsFromJSON(input io.Reader) (Playgrounds, error) {
	var playgrounds Playgrounds
	err := json.NewDecoder(input).Decode(&playgrounds)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse input %q into slice, '%v'", input, err)
	}
	return playgrounds, nil
}

func (p Playgrounds) Find(ID int) (Playground, int, error) {
	for index, playground := range p {
		if playground.ID == ID {
			return playground, index, nil
		}
	}
	return Playground{}, 0, ErrorNotFoundPlayground
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
		return strings.ToLower(p[i].Name) < strings.ToLower(p[j].Name)
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

func (p *Playground) AddComment(comment Comment) error {
	content := strings.TrimSpace(comment.Content)
	author := strings.TrimSpace(comment.Author)
	if content == "" || author == "" {
		return ErrEmptyField
	}
	newComment := Comment{
		Content:          comment.Content,
		Author:           comment.Author,
		ID:               len(p.Comments) + 1,
		TimeOfSubmission: comment.TimeOfSubmission,
	}
	p.Comments = append(Comments{newComment}, p.Comments...)
	return nil
}

func (p *Playground) DeleteComment(commentID int) error {
	for index, comment := range p.Comments {
		if comment.ID == commentID {
			p.Comments = append(p.Comments[:index], p.Comments[index+1:]...)
			return nil
		}
	}
	return errors.New("Couldn't find comment")
}
