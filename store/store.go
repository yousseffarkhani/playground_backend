package store

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrorParsingJson = errors.New("Couldn't parse file into JSON")
)

type PlaygroundStore interface {
	AllPlaygrounds() Playgrounds
	Playground(ID int) (Playground, error)
	NewPlayground(newPlayground Playground) map[string]error
	// AddComment(playgroundID int, comment Comment)
}

type Database struct {
	MainPlaygroundStore      PlaygroundStore
	SubmittedPlaygroundStore *SubmittedPlaygroundStore
}

type MainPlaygroundStore struct {
	playgrounds Playgrounds
}

type SubmittedPlaygroundStore struct {
	playgrounds Playgrounds
}

func NewFromFile(path string) (*MainPlaygroundStore, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("Problem opening %s, %s", file.Name(), err)
	}

	mainPlaygroundStore, err := New(file)
	if err != nil {
		return nil, fmt.Errorf("Problem creating database, %s", err)
	}

	return mainPlaygroundStore, nil
}

func New(file *os.File) (*MainPlaygroundStore, error) {
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
	for i, _ := range playgrounds {
		playgrounds[i].ID = i + 1
	}
	return &MainPlaygroundStore{playgrounds: playgrounds}, nil
}

func (s *MainPlaygroundStore) AllPlaygrounds() Playgrounds {
	s.playgrounds.sortByName()
	return s.playgrounds
}

func (s *SubmittedPlaygroundStore) AllPlaygrounds() Playgrounds {
	s.playgrounds.sortByName()
	return s.playgrounds
}

func (s *MainPlaygroundStore) Playground(ID int) (Playground, error) {
	playground, err := s.playgrounds.Find(ID)
	if err != nil {
		return Playground{}, ErrorNotFoundPlayground
	}
	return playground, nil
}

func (s *SubmittedPlaygroundStore) Playground(ID int) (Playground, error) {
	playground, err := s.playgrounds.Find(ID)
	if err != nil {
		return Playground{}, ErrorNotFoundPlayground
	}
	return playground, nil
}

func (s *MainPlaygroundStore) NewPlayground(newPlayground Playground) map[string]error {
	errorsMap := verifyCorrectPlaygroundInput(newPlayground)
	if len(errorsMap) > 0 {
		return errorsMap
	}
	if s.isAlreadyExisting(newPlayground) {
		errorsMap["Playground"] = errors.New("This playground already exists")
		return errorsMap
	}
	newPlayground.ID = len(s.playgrounds) + 1
	s.playgrounds = append(s.playgrounds, newPlayground)
	return nil
}

func (d *Database) SubmitPlayground(newPlayground Playground) map[string]error {
	errorsMap := verifyCorrectPlaygroundInput(newPlayground)
	if len(errorsMap) > 0 {
		return errorsMap
	}
	if isNameOrAddressAlreadyExisting(newPlayground, d.SubmittedPlaygroundStore.AllPlaygrounds()) {
		errorsMap["Playground"] = errors.New("This playground already exists")
		return errorsMap
	}
	if isNameOrAddressAlreadyExisting(newPlayground, d.MainPlaygroundStore.AllPlaygrounds()) {
		errorsMap["Playground"] = errors.New("This playground already exists")
		return errorsMap
	}
	newPlayground.ID = len(d.SubmittedPlaygroundStore.playgrounds) + 1
	d.SubmittedPlaygroundStore.playgrounds = append(d.SubmittedPlaygroundStore.playgrounds, newPlayground)
	return nil
}

func isNameOrAddressAlreadyExisting(newPlayground Playground, playgrounds Playgrounds) bool {
	for _, playground := range playgrounds {
		if strings.ToLower(playground.Name) == strings.ToLower(newPlayground.Name) {
			return true
		}
		if strings.ToLower(playground.Address) == strings.ToLower(newPlayground.Address) {
			return true
		}
	}
	return false
}

func (s *MainPlaygroundStore) isAlreadyExisting(newPlayground Playground) bool {
	if isNameOrAddressAlreadyExisting(newPlayground, s.AllPlaygrounds()) {
		return true
	}
	for _, playground := range s.AllPlaygrounds() {
		if playground.Long == newPlayground.Long && playground.Lat == newPlayground.Lat {
			return true
		}
	}
	return false
}

var ErrEmptyField = errors.New("Empty field")

func verifyCorrectPlaygroundInput(newPlayground Playground) map[string]error {
	errorsMap := make(map[string]error)
	value := reflect.ValueOf(newPlayground)
	typeOfData := value.Type()
	if value.Kind() == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			fieldName := typeOfData.Field(i).Name
			fieldValue := value.Field(i).String()
			if fieldName != "Coating" && fieldName != "Open" && fieldName != "Type" && strings.TrimSpace(fieldValue) == "" {
				errorsMap[fieldName] = ErrEmptyField
				continue
			}
			if fieldName == "PostalCode" {
				if _, err := strconv.Atoi(fieldValue); err != nil {
					errorsMap[fieldName] = errors.New("Postal Code should be a number")
					continue
				}
				if len(fieldValue) != 5 {
					errorsMap[fieldName] = errors.New("Postal Code should be 5 characters long")
				}
			}
		}
	}
	return errorsMap
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
