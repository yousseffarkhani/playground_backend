package psql

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/yousseffarkhani/playground/backend2/server"
)

func NewPlaygroundsFromJSON(input io.Reader) (server.Playgrounds, error) {
	var playgrounds server.Playgrounds
	err := json.NewDecoder(input).Decode(&playgrounds)
	if err != nil {
		return nil, ErrorParsingJson
	}
	return playgrounds, nil
}

func (db *playgroundDatabase) GetAllPlaygrounds() server.Playgrounds {
	var playgrounds server.Playgrounds
	db.Where("draft=?", false).Find(&playgrounds)
	return playgrounds
}

func (db *playgroundDatabase) GetPlayground(ID int) server.Playground {
	var playground server.Playground
	db.Where("id=? AND draft=?", ID, false).Find(&playground)
	return playground
}

func (db *playgroundDatabase) AddPlayground(newPlayground server.Playground) error {
	if err := db.Save(&newPlayground).Error; err != nil {
		return fmt.Errorf("Couldn't update comment, %s", err)
	}
	log.Println("New playground successfully created")
	return nil
}

func (db *playgroundDatabase) GetPlaygroundByName(name string) (server.Playground, error) {
	var playground server.Playground
	db.Where("name=?", name).Find(&playground)
	if playground.Name == "" {
		return server.Playground{}, ErrorNotFoundPlayground
	}
	return playground, nil
}
