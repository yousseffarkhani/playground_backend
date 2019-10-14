package store

import (
	"errors"
	"fmt"
	"os"
)

var (
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

	playgrounds, err := NewPlaygrounds(file)
	if err != nil {
		return nil, ErrorParsingJson
	}

	playgrounds.sortByName()
	return &simplePlaygroundStore{playgrounds}, nil
}

func (s *simplePlaygroundStore) AllPlaygrounds() Playgrounds {
	return s.playgrounds
}

func (s *simplePlaygroundStore) Playground(ID int) (Playground, error) {
	playground, err := s.playgrounds.Find(ID)
	if err != nil {
		return Playground{}, ErrorNotFoundPlayground
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
