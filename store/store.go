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
	for i, _ := range playgrounds {
		playgrounds[i].ID = i + 1
	}
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

func (s *simplePlaygroundStore) NewPlayground(newPlayground Playground) map[string]error {
	errorsMap := make(map[string]error)
	errorsMap = verifyCorrectPlaygroundInput(newPlayground)
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

func (s *simplePlaygroundStore) isAlreadyExisting(NewPlayground Playground) bool {
	for _, playground := range s.AllPlaygrounds() {
		if strings.ToLower(playground.Name) == strings.ToLower(NewPlayground.Name) {
			return true
		}
		if strings.ToLower(playground.Address) == strings.ToLower(NewPlayground.Address) {
			return true
		}
		if playground.Long == NewPlayground.Long && playground.Lat == NewPlayground.Lat {
			return true
		}
	}
	return false
}

var ErrEmptyField = errors.New("Empty field")

func verifyCorrectPlaygroundInput(NewPlayground Playground) map[string]error {
	errorsMap := make(map[string]error)
	value := reflect.ValueOf(NewPlayground)
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
