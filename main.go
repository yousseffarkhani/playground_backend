package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/yousseffarkhani/playground/backend2/authentication"
	"github.com/yousseffarkhani/playground/backend2/geolocationClient"
	"github.com/yousseffarkhani/playground/backend2/middleware"
	"github.com/yousseffarkhani/playground/backend2/psql"
	"github.com/yousseffarkhani/playground/backend2/views"

	"github.com/yousseffarkhani/playground/backend2/configuration"

	"github.com/yousseffarkhani/playground/backend2/server"
)

const (
	dbFileName = "playgroundsOpenData.json"
)

func init() {
	configuration.LoadEnvVariables()
	authentication.InitAuthentication()
}

func main() {
	var database server.Database
	var err error
	if database, err = psql.ExistingDatabase(); err != nil {
		database, err = psql.NewPlaygroundDatabaseFromFilepath(dbFileName)
		if err != nil {
			log.Fatalf("Problem opening %s %v", dbFileName, err)
		}
	}
	geolocationClient := &geolocationClient.APIGouvFR{}
	views := views.Initialize()
	middlewares := middleware.Initialize()
	svr := server.New(database, geolocationClient, views, middlewares)
	listenAndServe(svr)
}

func listenAndServe(svr *server.PlaygroundServer) {
	var port string
	if configuration.Variables.ProductionMode {
		port = ":443"
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Couldn't get working directory, %s", err)
		}

		log.Println("Listening on port", port)
		log.Fatal(http.ListenAndServeTLS(port, filepath.Join(pwd, configuration.Variables.TLS.PathToCertFile), filepath.Join(pwd, configuration.Variables.TLS.PathToPrivKey), svr))
		return
	}
	port = ":5000"
	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(port, svr))
}
