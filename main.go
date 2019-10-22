package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yousseffarkhani/playground/backend2/views"

	"github.com/yousseffarkhani/playground/backend2/geolocationClient"

	"github.com/yousseffarkhani/playground/backend2/store"

	"github.com/yousseffarkhani/playground/backend2/middleware"
	"github.com/yousseffarkhani/playground/backend2/server"
)

const (
	port       = ":443"
	dbFileName = "playgrounds.json"
)

func main() {
	database, err := store.NewFromFile(dbFileName)
	if err != nil {
		log.Fatalf("Problem opening %s %v", dbFileName, err)
	}
	geolocationClient := &geolocationClient.APIGouvFR{}
	views := views.Initialize()
	middlewares := middleware.Initialize()
	svr := server.New(database, geolocationClient, views, middlewares)
	pwd, _ := os.Getwd()
	pathToCertFile := os.Getenv("CERTFILE")
	pathToPrivKey := os.Getenv("PRIVKEY")
	fmt.Println("Listening on port", port)
	log.Fatal(http.ListenAndServeTLS(port, filepath.Join(pwd, pathToCertFile), filepath.Join(pwd, pathToPrivKey), svr))
}
