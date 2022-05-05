package main

import (
	"log"
	"os"

	"github.com/johannaojeling/go-rest-api/pkg/api"
)

func main() {
	driver, ok := os.LookupEnv("DRIVER")
	if !ok {
		driver = "postgres"
	}

	dbUrl := os.Getenv("DB_URL")
	app, err := api.NewApp(driver, dbUrl)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
		log.Printf("defaulting to port %s\n", port)
	}

	log.Printf("listening on port %s\n", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("error running app: %v", err)
	}
}
