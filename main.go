package main

import (
	"log"
	"os"

	"github.com/johannaojeling/go-rest-api/pkg/api"
)

var (
	driver = os.Getenv("DRIVER")
	dbUrl  = os.Getenv("DB_URL")
	port   = os.Getenv("PORT")
)

func init() {
	if driver == "" {
		driver = "postgres"
	}
	if port == "" {
		port = "8080"
	}
}

func main() {
	app, err := api.NewApp(driver, dbUrl)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	log.Printf("listening on port %s\n", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("error running app: %v", err)
	}
}
