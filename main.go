package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yousseffarkhani/playground/backend2/views"

	"github.com/yousseffarkhani/playground/backend2/geolocationClient"

	"github.com/yousseffarkhani/playground/backend2/store"

	"github.com/yousseffarkhani/playground/backend2/server"
)

const (
	port       = ":5000"
	dbFileName = "playgrounds.json"
)

func main() {
	database, err := store.NewFromFile(dbFileName)
	if err != nil {
		log.Fatalf("Problem opening %s %v", dbFileName, err)
	}
	geolocationClient := &geolocationClient.APIGouvFR{}
	views := views.Initialize()
	svr := server.New(database, geolocationClient, views)
	fmt.Println("Listening on port", port)
	http.ListenAndServe(port, svr)
}
