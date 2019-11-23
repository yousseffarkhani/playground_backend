package psql

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/yousseffarkhani/playground/backend2/configuration"
	"github.com/yousseffarkhani/playground/backend2/server"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

var (
	ErrorParsingJson               = errors.New("Couldn't parse file into JSON")
	ErrorNotFoundPlayground        = errors.New("server.Playground doesn't exist")
	ErrorPlaygroundAlreadyExisting = errors.New("This playground already exists")
)

var (
	driverName = "postgres"
	host       = "localhost"
	port       = "5432"
	user       string
	password   string
	dbname     = "basket"
)

type playgroundDatabase struct {
	*gorm.DB
}

func ExistingDatabase() (server.Database, error) {
	user = configuration.Variables.DBUser
	password = configuration.Variables.DBPassword
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable dbname=%s", host, port, user, password, dbname)
	gormDB, err := gorm.Open(driverName, psqlInfo)
	if err != nil {
		return nil, err
	}
	db := &playgroundDatabase{gormDB}
	return db, nil
}

func NewPlaygroundDatabaseFromFilepath(path string) (server.Database, error) {
	gormDB, err := initializeDB()
	if err != nil {
		return nil, fmt.Errorf("Problem initializing DB, %s", err)
	}
	db := &playgroundDatabase{gormDB}

	file, err := openPlaygroundsFile(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open file, %s", err)
	}
	defer file.Close()

	err = db.initData(file)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *playgroundDatabase) initData(file *os.File) error {
	playgrounds, err := NewPlaygroundsFromJSON(file)
	if err != nil {
		return fmt.Errorf("Problem loading playgrounds from file %s, %s", file.Name(), err)
	}

	playgrounds.SortByName()

	for i, playground := range playgrounds {
		playground.ID = i + 1
		playground.TimeOfSubmission = time.Now()
		playground.Author = "Admin"

		err := db.AddPlayground(playground)
		if err != nil {
			return err
		}
	}

	return nil
}

func initializeDB() (*gorm.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", host, port, user, password)
	db, err := gorm.Open(driverName, psqlInfo)
	if err != nil {
		return nil, err
	}
	err = reset(driverName, psqlInfo, dbname)
	if err != nil {
		return nil, err
	}
	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbname)
	db, err = gorm.Open(driverName, psqlInfo)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&server.Playground{})
	db.AutoMigrate(&server.Comment{})

	return db, nil
}

func reset(driverName, dataSource, dbname string) error {
	db, err := gorm.Open(driverName, dataSource)
	if err != nil {
		return err
	}
	resetDB(db, dbname)

	return db.Close()
}

func resetDB(db *gorm.DB, name string) {
	db.Exec("DROP DATABASE IF EXISTS " + name)
	createDB(db, name)
}

func createDB(db *gorm.DB, name string) {
	db.Exec("CREATE DATABASE " + name)
}

func (db *playgroundDatabase) close() error {
	return db.Close()
}

func initializeStoreFileIfEmpty(file *os.File) error {
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

// Opens the file. If file non existent creates a JSON file.
func openPlaygroundsFile(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Problem opening %s, %v", file.Name(), err)
	}

	err = initializeStoreFileIfEmpty(file)
	if err != nil {
		return nil, fmt.Errorf("Couldn't initialize %s, %v", file.Name(), err)
	}

	return file, nil
}
