package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/yousseffarkhani/playground/backend2/server"
)

var (
	// ErrorNotFoundPlayground = errors.New("Playground doesn't exist")
	ErrorParsingJson = errors.New("Couldn't parse file into JSON")
)

type simplePlaygroundStore struct {
	playgrounds Playgrounds
}

func NewFromFile(path string) (*simplePlaygroundStore, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("Problem opening %s, %s", file.Name(), err)
	}

	database, err := New(file)
	if err != nil {
		return nil, fmt.Errorf("Problem creating database, %s", err)
	}

	return database, nil
}

func New(file *os.File) (*simplePlaygroundStore, error) {
	defer file.Close()

	err := initializeStoreFile(file)
	if err != nil {
		return nil, fmt.Errorf("problem initialising db file, %v", err)
	}
	var playgrounds []server.Playground
	err = json.NewDecoder(file).Decode(&playgrounds)
	if err != nil {
		return nil, ErrorParsingJson
	}
	return &simplePlaygroundStore{playgrounds}, nil
}

func (s *simplePlaygroundStore) AllPlaygrounds() []server.Playground {
	sort.Slice(s.playgrounds, func(i, j int) bool {
		return s.playgrounds[i].Name < s.playgrounds[j].Name
	})
	return s.playgrounds
}

func (s *simplePlaygroundStore) Playground(ID int) (server.Playground, error) {
	playground, err := s.playgrounds.Find(ID)
	if err != nil {
		return server.Playground{}, ErrorNotFoundPlayground
	}
	return playground, nil
}

func initializeStoreFile(file *os.File) error {
	file.Seek(0, 0)
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Couldn't get file informations, %s", err)
	}
	if fileInfo.Size() == 0 {
		file.Write([]byte("[]"))
		file.Seek(0, 0)
	}
	return nil
}
